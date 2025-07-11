package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"yambot/pkg/config"
)

func (b *Bot) registerCommands() error {
	log.Printf("Registering %d commands...", len(b.Config.GetCommands()))

	for _, cmd := range b.Config.GetCommands() {
		if err := b.registerCommand(cmd); err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		log.Printf("Registered command: %s (type: %s)", cmd.Name, cmd.Type)
	}

	return nil
}

func (b *Bot) registerCommand(cmd config.CommandSpec) error {
	switch cmd.Type {
	case "slash":
		return b.registerSlashCommand(cmd)
	case "modal":
		return b.registerModalCommand(cmd)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

func (b *Bot) registerSlashCommand(cmd config.CommandSpec) error {
	options := make([]*discordgo.ApplicationCommandOption, 0)

	for _, field := range cmd.Fields {
		option, err := b.createCommandOption(field)
		if err != nil {
			return fmt.Errorf("failed to create option for field %s: %w", field.Name, err)
		}
		options = append(options, option)
	}

	command := &discordgo.ApplicationCommand{
		Name:        cmd.Name,
		Description: fmt.Sprintf("Execute %s command", cmd.Name),
		Options:     options,
	}

	_, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, "", command)
	if err != nil {
		return fmt.Errorf("failed to create slash command: %w", err)
	}

	return nil
}

func (b *Bot) registerModalCommand(cmd config.CommandSpec) error {
	command := &discordgo.ApplicationCommand{
		Name:        cmd.Name,
		Description: fmt.Sprintf("Open %s form", cmd.Name),
		Options:     []*discordgo.ApplicationCommandOption{},
	}

	_, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, "", command)
	if err != nil {
		return fmt.Errorf("failed to create modal command: %w", err)
	}

	return nil
}

func (b *Bot) createCommandOption(field config.FieldSpec) (*discordgo.ApplicationCommandOption, error) {
	var optionType discordgo.ApplicationCommandOptionType
	var choices []*discordgo.ApplicationCommandOptionChoice

	switch field.Type {
	case "text":
		optionType = discordgo.ApplicationCommandOptionString
	case "select":
		optionType = discordgo.ApplicationCommandOptionString
		if len(field.Options) > 0 {
			choices = make([]*discordgo.ApplicationCommandOptionChoice, 0)
			for _, option := range field.Options {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  option,
					Value: option,
				})
			}
		}
	case "attachment":
		optionType = discordgo.ApplicationCommandOptionAttachment
	default:
		return nil, fmt.Errorf("unsupported field type: %s", field.Type)
	}

	return &discordgo.ApplicationCommandOption{
		Type:        optionType,
		Name:        field.Name,
		Description: fmt.Sprintf("Enter %s", field.Name),
		Required:    true,
		Choices:     choices,
	}, nil
}
