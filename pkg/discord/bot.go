package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"yambot/pkg/config"
)

type Bot struct {
	Session   *discordgo.Session
	Config    *config.Config
	Templater *ResponseTemplater
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
		Session:   session,
		Config:    cfg,
		Templater: NewResponseTemplater(),
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
	b.dispatchCommand(s, i)
}

func (b *Bot) dispatchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if this is an application command interaction
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	
	if i.ApplicationCommandData().Name == "" {
		log.Printf("Received interaction with empty command name")
		return
	}

	commandName := i.ApplicationCommandData().Name
	log.Printf("Dispatching command: %s", commandName)

	commandSpec := b.findCommandSpec(commandName)
	if commandSpec == nil {
		log.Printf("Unknown command: %s", commandName)
		b.respondWithError(s, i, "Error: Unknown command")
		return
	}

	log.Printf("Found command spec for %s (type: %s)", commandName, commandSpec.Type)

	err := b.routeToHandler(s, i, commandSpec)
	if err != nil {
		log.Printf("Error handling command %s: %v", commandName, err)
		b.respondWithError(s, i, "Error: Internal error processing command")
	}
}

func (b *Bot) findCommandSpec(commandName string) *config.CommandSpec {
	for _, cmd := range b.Config.GetCommands() {
		if cmd.Name == commandName {
			return &cmd
		}
	}
	return nil
}

func (b *Bot) routeToHandler(s *discordgo.Session, i *discordgo.InteractionCreate, commandSpec *config.CommandSpec) error {
	switch commandSpec.Type {
	case "slash":
		return b.handleSlashCommand(s, i, commandSpec)
	case "modal":
		return b.handleModalCommand(s, i, commandSpec)
	default:
		return fmt.Errorf("unknown command type: %s", commandSpec.Type)
	}
}

func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmd *config.CommandSpec) error {
	options := i.ApplicationCommandData().Options

	// Extract input data
	inputs := make(map[string]interface{})
	for _, option := range options {
		switch option.Type {
		case discordgo.ApplicationCommandOptionString:
			inputs[option.Name] = option.StringValue()
		case discordgo.ApplicationCommandOptionAttachment:
			attachmentID := option.Value.(string)
			if attachment, exists := i.ApplicationCommandData().Resolved.Attachments[attachmentID]; exists {
				inputs[option.Name] = map[string]interface{}{
					"name":         attachment.Filename,
					"content_type": attachment.ContentType,
					"size":         attachment.Size,
					"url":          attachment.URL,
				}
			} else {
				inputs[option.Name] = attachmentID
			}
		default:
			inputs[option.Name] = option.Value
		}
	}

	// Prepare template data
	templateData := &TemplateData{
		Inputs: inputs,
	}

	// Send webhook if specified
	if cmd.Webhook != "" {
		webhookResp := b.sendSlashWebhook(cmd.Webhook, inputs)
		templateData.WebhookResponse = webhookResp
	}

	// Generate response using template
	var response string
	if cmd.ResponseFormat != "" {
		response = b.Templater.RenderWithFallback(cmd.ResponseFormat, templateData)
	} else {
		response = "✅ Komenda została przyjęta."
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	if err != nil {
		return fmt.Errorf("error responding to slash command interaction: %w", err)
	}

	return nil
}

func (b *Bot) handleModalCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmd *config.CommandSpec) error {
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
		return fmt.Errorf("error responding with modal: %w", err)
	}

	return nil
}

func (b *Bot) fetchRemoteOptions(webhookURL string) ([]config.RemoteOption, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", webhookURL, nil)
	if err != nil {
		log.Printf("Error creating request for remote options: %v", err)
		return nil, fmt.Errorf("failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching remote options from %s: %v", webhookURL, err)
		return nil, fmt.Errorf("failed to fetch remote options")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Remote options webhook returned non-success status %d for URL %s", resp.StatusCode, webhookURL)
		return nil, fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	var options []config.RemoteOption

	// Try to decode as expected format first
	if err := json.NewDecoder(resp.Body).Decode(&options); err != nil {
		log.Printf("Failed to decode as RemoteOption array, trying alternative formats: %v", err)

		// Reset response body for another attempt
		resp.Body.Close()
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to re-fetch remote options")
		}
		defer resp.Body.Close()

		// Try to decode as generic array of objects
		var genericOptions []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&genericOptions); err != nil {
			log.Printf("Failed to decode as generic array: %v", err)
			return nil, fmt.Errorf("failed to decode response")
		}

		// Convert to RemoteOption format
		options = make([]config.RemoteOption, 0, len(genericOptions))
		for _, item := range genericOptions {
			label := ""
			value := ""

			// Try common field names
			if name, ok := item["name"].(string); ok {
				label = name
				value = name
			} else if title, ok := item["title"].(string); ok {
				label = title
				value = title
			} else if id, ok := item["id"]; ok {
				label = fmt.Sprintf("ID: %v", id)
				value = fmt.Sprintf("%v", id)
			}

			if label != "" && value != "" {
				options = append(options, config.RemoteOption{
					Label: label,
					Value: value,
				})
			}
		}
	}

	log.Printf("Successfully fetched %d remote options from %s", len(options), webhookURL)
	return options, nil
}

func (b *Bot) sendSlashWebhook(webhookURL string, inputs map[string]interface{}) *WebhookResponse {
	payload, err := json.Marshal(inputs)
	if err != nil {
		log.Printf("Error marshaling inputs for webhook: %v", err)
		return &WebhookResponse{
			StatusCode: 0,
			Status:     "error",
			Error:      "Failed to prepare webhook data",
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating webhook request: %v", err)
		return &WebhookResponse{
			StatusCode: 0,
			Status:     "error",
			Error:      "Failed to create webhook request",
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending webhook to %s: %v", webhookURL, err)
		return &WebhookResponse{
			StatusCode: 0,
			Status:     "error",
			Error:      "Failed to send webhook",
		}
	}
	defer resp.Body.Close()

	webhookResponse := &WebhookResponse{
		StatusCode: resp.StatusCode,
		Status:     fmt.Sprintf("%d", resp.StatusCode),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		webhookResponse.Status = "success"
		
		// Try to decode response body as JSON
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err == nil {
			webhookResponse.Data = responseData
		}
	} else {
		webhookResponse.Status = "error"
		webhookResponse.Error = fmt.Sprintf("Webhook returned status %d", resp.StatusCode)
	}

	log.Printf("Webhook sent to %s, status: %s", webhookURL, webhookResponse.Status)
	return webhookResponse
}
