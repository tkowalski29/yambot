package main

import (
	"log"
	"os"

	"yambot/pkg/config"
	"yambot/pkg/discord"
)

func main() {
	configPath := "cmd/config.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Successfully loaded configuration with %d commands", len(cfg.Commands))

	bot, err := discord.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}

	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
