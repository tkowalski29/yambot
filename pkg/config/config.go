package config

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot      BotConfig     `yaml:"bot"`
	Commands []CommandSpec `yaml:"commands"`
}

type BotConfig struct {
	Discord DiscordConfig `yaml:"discord"`
}

type DiscordConfig struct {
	Token string `yaml:"token"`
}

type CommandSpec struct {
	Name           string      `yaml:"name"`
	Type           string      `yaml:"type"`
	Webhook        string      `yaml:"webhook"`
	Fields         []FieldSpec `yaml:"fields"`
	ResponseFormat string      `yaml:"response_format,omitempty"`
}

type FieldSpec struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Options  []string `yaml:"options,omitempty"`
	Webhook  string   `yaml:"webhook,omitempty"`
	Required bool     `yaml:"required,omitempty"`
}

type RemoteOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate templates in commands
	if err := config.ValidateTemplates(); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	return &config, nil
}

func (c *Config) GetCommands() []CommandSpec {
	return c.Commands
}

func (c *Config) GetDiscordToken() string {
	return c.Bot.Discord.Token
}

func (c *Config) ValidateTemplates() error {
	for _, cmd := range c.Commands {
		if cmd.ResponseFormat != "" {
			// Create template with the same custom functions as in templating.go
			_, err := template.New("validation").Funcs(template.FuncMap{
				"upper": strings.ToUpper,
				"lower": strings.ToLower,
				"title": strings.Title,
				"json": func(v interface{}) string {
					return fmt.Sprintf("%v", v) // Simplified for validation
				},
			}).Parse(cmd.ResponseFormat)
			if err != nil {
				return fmt.Errorf("invalid template in command '%s': %w", cmd.Name, err)
			}
		}
	}
	return nil
}
