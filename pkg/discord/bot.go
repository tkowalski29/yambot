package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"yambot/pkg/config"
)

type Bot struct {
	Session *discordgo.Session
	Config  *config.Config
}

func NewBot(cfg *config.Config) (*Bot, error) {
	token := cfg.GetDiscordToken()
	if token == "" {
		return nil, fmt.Errorf("discord token is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	return &Bot{
		Session: session,
		Config:  cfg,
	}, nil
}

func (b *Bot) Start() error {
	b.Session.AddHandler(b.handleInteraction)
	b.Session.AddHandler(b.handleModalSubmit)

	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")

	err = b.registerCommands()
	if err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot...")
	return b.Session.Close()
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	commandName := i.ApplicationCommandData().Name
	log.Printf("Received interaction for command: %s", commandName)

	var commandSpec *config.CommandSpec
	for _, cmd := range b.Config.GetCommands() {
		if cmd.Name == commandName {
			commandSpec = &cmd
			break
		}
	}

	if commandSpec == nil {
		log.Printf("Unknown command: %s", commandName)
		return
	}

	switch commandSpec.Type {
	case "slash":
		b.handleSlashCommand(s, i, commandSpec)
	case "modal":
		b.handleModalCommand(s, i, commandSpec)
	default:
		log.Printf("Unknown command type: %s", commandSpec.Type)
	}
}

func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmd *config.CommandSpec) {
	options := i.ApplicationCommandData().Options

	response := fmt.Sprintf("Received slash command: %s\n", cmd.Name)

	if len(options) > 0 {
		response += "\nSubmitted data:"
		for _, option := range options {
			switch option.Type {
			case discordgo.ApplicationCommandOptionString:
				response += fmt.Sprintf("\n**%s**: %s", option.Name, option.StringValue())
			case discordgo.ApplicationCommandOptionAttachment:
				attachmentID := option.Value.(string)
				if attachment, exists := i.ApplicationCommandData().Resolved.Attachments[attachmentID]; exists {
					response += fmt.Sprintf("\n**%s**: %s (%s, %d bytes)", option.Name, attachment.Filename, attachment.ContentType, attachment.Size)
				} else {
					response += fmt.Sprintf("\n**%s**: %s", option.Name, attachmentID)
				}
			default:
				response += fmt.Sprintf("\n**%s**: %v", option.Name, option.Value)
			}
		}
	}

	if cmd.Webhook != "" {
		response += fmt.Sprintf("\n\nData will be sent to webhook: %s", cmd.Webhook)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

func (b *Bot) handleModalCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmd *config.CommandSpec) {
	components := b.createModalComponents(cmd)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   fmt.Sprintf("modal_%s", cmd.Name),
			Title:      fmt.Sprintf("Form: %s", cmd.Name),
			Components: components,
		},
	})

	if err != nil {
		log.Printf("Error responding with modal: %v", err)
	}
}
