package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testConfig := `bot:
  discord:
    token: TEST_TOKEN

commands:
  - name: test-command
    type: slash
    webhook: "https://example.com/webhook"
    fields:
      - name: field1
        type: text
      - name: field2
        type: select
        options: ["option1", "option2"]`

	tmpFile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Bot.Discord.Token != "TEST_TOKEN" {
		t.Errorf("Expected token 'TEST_TOKEN', got '%s'", cfg.Bot.Discord.Token)
	}

	if len(cfg.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(cfg.Commands))
	}

	cmd := cfg.Commands[0]
	if cmd.Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", cmd.Name)
	}

	if cmd.Type != "slash" {
		t.Errorf("Expected command type 'slash', got '%s'", cmd.Type)
	}

	if cmd.Webhook != "https://example.com/webhook" {
		t.Errorf("Expected webhook 'https://example.com/webhook', got '%s'", cmd.Webhook)
	}

	if len(cmd.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(cmd.Fields))
	}

	field1 := cmd.Fields[0]
	if field1.Name != "field1" || field1.Type != "text" {
		t.Errorf("Expected field1 (text), got %s (%s)", field1.Name, field1.Type)
	}

	field2 := cmd.Fields[1]
	if field2.Name != "field2" || field2.Type != "select" {
		t.Errorf("Expected field2 (select), got %s (%s)", field2.Name, field2.Type)
	}

	if len(field2.Options) != 2 {
		t.Errorf("Expected 2 options for field2, got %d", len(field2.Options))
	}
}

func TestGetCommands(t *testing.T) {
	cfg := &Config{
		Commands: []CommandSpec{
			{Name: "cmd1", Type: "slash"},
			{Name: "cmd2", Type: "modal"},
		},
	}

	commands := cfg.GetCommands()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}

func TestGetDiscordToken(t *testing.T) {
	cfg := &Config{
		Bot: BotConfig{
			Discord: DiscordConfig{
				Token: "TEST_TOKEN",
			},
		},
	}

	token := cfg.GetDiscordToken()
	if token != "TEST_TOKEN" {
		t.Errorf("Expected token 'TEST_TOKEN', got '%s'", token)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	invalidYAML := `invalid: yaml: content: [unclosed`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestValidateTemplates(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		hasError  bool
		errorMsg  string
	}{
		{
			name: "valid templates",
			config: &Config{
				Commands: []CommandSpec{
					{
						Name:           "test1",
						ResponseFormat: "Hello {{ .Inputs.name }}!",
					},
					{
						Name:           "test2",
						ResponseFormat: "Status: {{ .webhook_response.status }}",
					},
				},
			},
			hasError: false,
		},
		{
			name: "empty templates are valid",
			config: &Config{
				Commands: []CommandSpec{
					{
						Name:           "test1",
						ResponseFormat: "",
					},
					{
						Name: "test2",
						// No ResponseFormat field
					},
				},
			},
			hasError: false,
		},
		{
			name: "invalid template syntax",
			config: &Config{
				Commands: []CommandSpec{
					{
						Name:           "test1",
						ResponseFormat: "Hello {{ .Inputs.name",
					},
				},
			},
			hasError: true,
			errorMsg: "invalid template in command 'test1'",
		},
		{
			name: "complex valid template",
			config: &Config{
				Commands: []CommandSpec{
					{
						Name: "test1",
						ResponseFormat: `{{ if .webhook_response }}
Status: {{ .webhook_response.status }}
{{ if .webhook_response.data }}
ID: {{ .webhook_response.data.id }}
{{ end }}
{{ else }}
No webhook response
{{ end }}`,
					},
				},
			},
			hasError: false,
		},
		{
			name: "template with functions",
			config: &Config{
				Commands: []CommandSpec{
					{
						Name:           "test1",
						ResponseFormat: "Name: {{ upper .Inputs.name }}, Email: {{ lower .Inputs.email }}",
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateTemplates()
			
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLoadConfigWithTemplateValidation(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		expectError  bool
		errorContains string
	}{
		{
			name: "valid config with template",
			configYAML: `bot:
  discord:
    token: TEST_TOKEN

commands:
  - name: test-command
    type: slash
    webhook: "https://example.com/webhook"
    response_format: "Hello {{ .Inputs.name }}!"
    fields:
      - name: name
        type: text`,
			expectError: false,
		},
		{
			name: "invalid template in config",
			configYAML: `bot:
  discord:
    token: TEST_TOKEN

commands:
  - name: test-command
    type: slash
    webhook: "https://example.com/webhook"
    response_format: "Hello {{ .Inputs.name"
    fields:
      - name: name
        type: text`,
			expectError:   true,
			errorContains: "template validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-config-*.yml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.configYAML); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			_, err = LoadConfig(tmpFile.Name())
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
