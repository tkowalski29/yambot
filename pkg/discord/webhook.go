package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

// WebhookService handles webhook operations
type WebhookService struct{}

// NewWebhookService creates a new webhook service instance
func NewWebhookService() *WebhookService {
	return &WebhookService{}
}

// SendWebhook sends form data to a webhook URL
func (ws *WebhookService) SendWebhook(webhookURL string, formData map[string]string) error {
	payload, err := json.Marshal(formData)
	if err != nil {
		log.Printf("Error marshaling form data for webhook: %v", err)
		return fmt.Errorf("failed to prepare webhook data")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating webhook request: %v", err)
		return fmt.Errorf("failed to create webhook request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending webhook to %s: %v", webhookURL, err)
		return fmt.Errorf("failed to send webhook")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook returned non-success status %d for URL %s", resp.StatusCode, webhookURL)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	log.Printf("Successfully sent webhook to %s", webhookURL)
	return nil
}

// SendSlashCommandWebhook sends slash command data to a webhook URL
func (ws *WebhookService) SendSlashCommandWebhook(webhookURL string, commandName string, options []*discordgo.ApplicationCommandInteractionDataOption, attachments map[string]*discordgo.MessageAttachment) error {
	// Convert slash command options to map
	formData := make(map[string]string)
	formData["command"] = commandName

	for _, option := range options {
		switch option.Type {
		case discordgo.ApplicationCommandOptionString:
			formData[option.Name] = option.StringValue()
		case discordgo.ApplicationCommandOptionAttachment:
			attachmentID := option.Value.(string)
			if attachment, exists := attachments[attachmentID]; exists {
				// For now, we'll include file info in the JSON payload
				// In a more advanced implementation, you might want to send multipart/form-data
				formData[option.Name] = fmt.Sprintf("File: %s (%s, %d bytes)", attachment.Filename, attachment.ContentType, attachment.Size)
				formData[option.Name+"_url"] = attachment.URL
				formData[option.Name+"_content_type"] = attachment.ContentType
				formData[option.Name+"_size"] = fmt.Sprintf("%d", attachment.Size)
			} else {
				formData[option.Name] = attachmentID
			}
		default:
			formData[option.Name] = fmt.Sprintf("%v", option.Value)
		}
	}

	return ws.SendWebhook(webhookURL, formData)
}

// downloadAttachment downloads a file from Discord URL
func (ws *WebhookService) downloadAttachment(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
