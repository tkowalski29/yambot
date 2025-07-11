package config

import (
	"fmt"
	"io"
	"os"

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
	Name    string      `yaml:"name"`
	Type    string      `yaml:"type"`
	Webhook string      `yaml:"webhook"`
	Fields  []FieldSpec `yaml:"fields"`
}

type FieldSpec struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Options  []string `yaml:"options,omitempty"`
	Required bool     `yaml:"required,omitempty"`
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

	return &config, nil
}

func (c *Config) GetCommands() []CommandSpec {
	return c.Commands
}

func (c *Config) GetDiscordToken() string {
	return c.Bot.Discord.Token
}
