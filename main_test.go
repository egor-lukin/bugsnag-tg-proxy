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

func TestCheckTrigger(t *testing.T) {
	tests := []struct {
		name           string
		triggerInfo    TriggerInfo
		triggersConfig TriggersConfig
		expected       Trigger
		expectError    bool
	}{
		{
			name: "Valid trigger type",
			triggerInfo: TriggerInfo{
				Type: "exception", // Updated to match the casing
			},
			triggersConfig: TriggersConfig{
				Exception: Trigger{Enabled: true, Template: "Template for exception"},
			},
			expected:    Trigger{Enabled: true, Template: "Template for exception"},
			expectError: false,
		},
		{
			name: "Invalid trigger type",
			triggerInfo: TriggerInfo{
				Type: "new_trigger_type",
			},
			triggersConfig: TriggersConfig{},
			expected:       Trigger{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checkTrigger(tt.triggerInfo, tt.triggersConfig)
			if (err != nil) != tt.expectError {
				t.Errorf("checkTrigger() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("checkTrigger() = %v, want %v", result, tt.expected)
			}
		})
	}
}
