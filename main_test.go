package main

import (
	"testing"
)

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name        string
		response    BugsnagResponse
		templateStr string
		expected    string
		expectError bool
	}{
		{
			name: "Valid template",
			response: BugsnagResponse{
				Account: Account{
					ID:   "56b9ca7f17025f8756f69054",
					Name: "My Company",
					URL:  "https://app.bugsnag.com/accounts/my-company",
				},
				Project: Project{
					ID:   "56b9ca7f17025f8756f69054",
					Name: "My Test Project",
					URL:  "https://app.bugsnag.com/my-company/my-test-project",
				},
				Trigger: TriggerInfo{
					Type:    "exception",
					Message: "1000th exception",
				},
				Error: Error{
					ID:             "56b9ca7f17025f8756f69054",
					ErrorID:        "56b9ca7f17025f8756f69054",
					ExceptionClass: "NoMethodError",
					Message:        "Unable to connect to database.",
					App: ErrorApp{
						ReleaseStage: "production",
						Type:         "my app",
					},
				},
			},
			templateStr: "[{{ .Error.App.ReleaseStage }}, {{ .Error.App.Type }}, {{ .Trigger.Message }}] Error: {{.Error.Message}} from {{.Account.Name}}",
			expected:    "[production, my app, 1000th exception] Error: Unable to connect to database. from My Company",
			expectError: false,
		},
		{
			name: "Invalid template",
			response: BugsnagResponse{
				Account: Account{
					ID:   "56b9ca7f17025f8756f69054",
					Name: "My Company",
					URL:  "https://app.bugsnag.com/accounts/my-company",
				},
			},
			templateStr: "Error: {{.InvalidField}}",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatMessage(tt.response, tt.templateStr)
			if (err != nil) != tt.expectError {
				t.Errorf("formatMessage() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("formatMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTakeTemplate(t *testing.T) {
	tests := []struct {
		name        string
		triggerInfo TriggerInfo
		config      Config
		expected    string
		expectError bool
	}{
		{
			name: "Valid trigger type and defined specific template",
			triggerInfo: TriggerInfo{
				Type: "exception", // Updated to match the casing
			},
			config: Config{
				Telegram: TelegramConfig{
					ChatID:   "chatId",
					ApiToken: "apiToken",
				},
				Triggers: TriggersConfig{
					Exception: Trigger{Template: "specific template"},
				},
				CommonTemplate: "common template",
			},
			expected:    "specific template",
			expectError: false,
		},
		{
			name: "Valid trigger type and does not defined specific template",
			triggerInfo: TriggerInfo{
				Type: "exception",
			},
			config: Config{
				Telegram: TelegramConfig{
					ChatID:   "chatId",
					ApiToken: "apiToken",
				},
				Triggers: TriggersConfig{
					Exception: Trigger{Template: ""},
				},
				CommonTemplate: "common template",
			},
			expected:    "common template",
			expectError: false,
		},
		{
			name: "Invalid trigger type",
			triggerInfo: TriggerInfo{
				Type: "new_trigger_type",
			},
			config: Config{
				Telegram: TelegramConfig{
					ChatID:   "chatId",
					ApiToken: "apiToken",
				},
				Triggers: TriggersConfig{
					Exception: Trigger{Template: "specific template"},
				},
				CommonTemplate: "common template",
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := takeTemplate(tt.triggerInfo, tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("takeTemplate() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("takeTemplate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
