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
