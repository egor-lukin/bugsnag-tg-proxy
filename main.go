package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
	"unicode"
)

type TelegramConfig struct {
	ChatID   string `yaml:"chatId"`
	ApiToken string `yaml:"apiToken"`
}

type TriggersConfig struct {
	FirstException              Trigger `yaml:"firstException"`
	ErrorEventFrequency         Trigger `yaml:"errorEventFrequency"`
	PowerTen                    Trigger `yaml:"powerTen"`
	Exception                   Trigger `yaml:"exception"`
	Reopened                    Trigger `yaml:"reopened"`
	ProjectSpiking              Trigger `yaml:"projectSpiking"`
	Release                     Trigger `yaml:"release"`
	ProjectSamplingStarted      Trigger `yaml:"projectSamplingStarted"`
	ProjectApproachingRateLimit Trigger `yaml:"projectApproachingRateLimit"`
	Comment                     Trigger `yaml:"comment"`
	ErrorStateManualChange      Trigger `yaml:"errorStateManualChange"`
}

type Trigger struct {
	Enabled  bool   `yaml:"enabled"`
	Template string `yaml:"template,omitempty"`
}

type Config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Triggers TriggersConfig `yaml:"triggers"`
}

type BugsnagResponse struct {
	Account Account     `json:"account"`
	Project Project     `json:"project"`
	Trigger TriggerInfo `json:"trigger"`
	User    User        `json:"user"`
	Error   Error       `json:"error"`
	Release Release     `json:"release"`
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TriggerInfo struct {
	Type        string     `json:"type"`
	Message     string     `json:"message"`
	SnoozeRule  SnoozeRule `json:"snoozeRule"`
	Rate        int        `json:"rate,omitempty"`
	StateChange string     `json:"stateChange,omitempty"`
}

type SnoozeRule struct {
	Type      string `json:"type"`
	RuleValue int    `json:"ruleValue"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Error struct {
	ID                string   `json:"id"`
	ErrorID           string   `json:"errorId"`
	ExceptionClass    string   `json:"exceptionClass"`
	Message           string   `json:"message"`
	Context           string   `json:"context"`
	FirstReceived     string   `json:"firstReceived"`
	ReceivedAt        string   `json:"receivedAt"`
	RequestURL        string   `json:"requestUrl,omitempty"`
	AssignedUserID    string   `json:"assignedUserId,omitempty"`
	AssignedUserEmail string   `json:"assignedUserEmail,omitempty"`
	URL               string   `json:"url"`
	Severity          string   `json:"severity"`
	Status            string   `json:"status"`
	Unhandled         bool     `json:"unhandled"`
	App               ErrorApp `json:"app"`
}

type ErrorApp struct {
	ReleaseStage string `json:"releaseStage"`
	Type         string `json:"type"`
}

type Release struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	VersionCode   string `json:"versionCode,omitempty"`
	BundleVersion string `json:"bundleVersion,omitempty"`
	ReleaseStage  string `json:"releaseStage"`
	URL           string `json:"url"`
	ReleaseTime   string `json:"releaseTime"`
	ReleasedBy    string `json:"releasedBy,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        http.HandlerFunc(bugsnagWebhooksHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

func bugsnagWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/webhooks" {
		var jsonBody BugsnagResponse
		if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
			fmt.Printf("%+v\n", r.Body)
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		configPath := os.Getenv("CONFIG_PATH")
		if configPath == "" {
			configPath = "config.yaml"
		}

		config, err := readConfig(configPath)
		if err != nil {
			fmt.Printf("%+v\n", err)
			http.Error(w, "Failed to read config", http.StatusInternalServerError)
			return
		}

		trigger, err := checkTrigger(jsonBody.Trigger, config.Triggers)
		if err != nil {
			fmt.Printf("%+v\n", err)
			http.Error(w, "Invalid Config", http.StatusInternalServerError)
			return
		}

		if !trigger.Enabled {
			w.WriteHeader(http.StatusOK)
			// add log
			return
		}

		message, err := formatMessage(jsonBody, trigger.Template)
		if err != nil {
			http.Error(w, "Invalid template format", http.StatusInternalServerError)
			return
		}

		err = sendMessageToTelegram(message, config.Telegram)
		if err != nil {
			http.Error(w, "Failed to send message to Telegram", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.NotFound(w, r)
}

func checkTrigger(triggerInfo TriggerInfo, triggersConfig TriggersConfig) (Trigger, error) {
	validTriggerTypes := map[string]string{
		"FirstException":              "FirstException",
		"ErrorEventFrequency":         "ErrorEventFrequency",
		"PowerTen":                    "PowerTen",
		"Exception":                   "Exception",
		"Reopened":                    "Reopened",
		"ProjectSpiking":              "ProjectSpiking",
		"Release":                     "Release",
		"ProjectSamplingStarted":      "ProjectSamplingStarted",
		"ProjectApproachingRateLimit": "ProjectApproachingRateLimit",
		"ErrorStateManualChange":      "ErrorStateManualChange",
	}

	triggerType := capitalizeFirstLetter(triggerInfo.Type)
	if validType, exists := validTriggerTypes[triggerType]; exists {
		trigger := reflect.ValueOf(triggersConfig).FieldByName(validType)
		if trigger.IsValid() {
			return trigger.Interface().(Trigger), nil
		}
	}

	return Trigger{}, fmt.Errorf("invalid trigger type: %s", triggerInfo.Type)
}

func capitalizeFirstLetter(str string) string {
	if len(str) == 0 {
		return str
	}

	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

func formatMessage(jsonBody BugsnagResponse, templateStr string) (string, error) {
	tmpl, err := template.New("message").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var messageBuffer bytes.Buffer
	err = tmpl.Execute(&messageBuffer, jsonBody)
	if err != nil {
		return "", err
	}

	return messageBuffer.String(), nil
}

func sendMessageToTelegram(message string, config TelegramConfig) error {
	messageData := map[string]interface{}{
		"chat_id": config.ChatID,
		"text":    message,
	}
	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		return err
	}

	resp, err := http.Post("https://api.telegram.org/bot"+config.ApiToken+"/sendMessage", "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func readConfig(filePath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}
