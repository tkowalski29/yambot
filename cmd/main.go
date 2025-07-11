package main

import (
	"fmt"
	"log"
	"os"

	"yambot/pkg/config"
)

func main() {
	configPath := "config.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Successfully loaded configuration with %d commands", len(cfg.Commands))
	
	fmt.Println("Available commands:")
	for _, cmd := range cfg.GetCommands() {
		fmt.Printf("- %s (type: %s)\n", cmd.Name, cmd.Type)
		fmt.Printf("  webhook: %s\n", cmd.Webhook)
		fmt.Printf("  fields:\n")
		for _, field := range cmd.Fields {
			fmt.Printf("    - %s (%s)", field.Name, field.Type)
			if len(field.Options) > 0 {
				fmt.Printf(" options: %v", field.Options)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}