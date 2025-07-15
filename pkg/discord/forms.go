package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"yambot/pkg/config"
)

func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if this is actually a modal submit interaction
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	
	if !strings.HasPrefix(i.ModalSubmitData().CustomID, "modal_") {
		return
	}

	commandName := strings.TrimPrefix(i.ModalSubmitData().CustomID, "modal_")
	log.Printf("Received modal submission for command: %s", commandName)

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

	modalData := i.ModalSubmitData()
	formData := b.extractFormData(&modalData)

	if err := b.validateFormData(commandSpec, formData); err != nil {
		response := fmt.Sprintf("❌ **Validation Error**\n\n%s\n\nPlease check your input and try again.", err.Error())
		b.respondWithError(s, i, response)
		return
	}

	var webhookResp *WebhookResponse
	if commandSpec.Webhook != "" {
		webhookResp = b.sendWebhook(commandSpec.Webhook, formData)
	}

	response := b.createFormResponse(commandSpec, formData, webhookResp)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	if err != nil {
		log.Printf("Error responding to modal submission: %v", err)
	}
}

func (b *Bot) extractFormData(data *discordgo.ModalSubmitInteractionData) map[string]string {
	formData := make(map[string]string)

	for _, component := range data.Components {
		if actionRow, ok := component.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if textInput, ok := comp.(*discordgo.TextInput); ok {
					formData[textInput.CustomID] = textInput.Value
				}
			}
		}
	}

	return formData
}

func (b *Bot) sendWebhook(webhookURL string, formData map[string]string) *WebhookResponse {
	// Convert formData to interface{} map for consistency
	inputs := make(map[string]interface{})
	for k, v := range formData {
		inputs[k] = v
	}

	payload, err := json.Marshal(inputs)
	if err != nil {
		log.Printf("Error marshaling form data for webhook: %v", err)
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

func (b *Bot) createFormResponse(cmd *config.CommandSpec, formData map[string]string, webhookResp *WebhookResponse) string {
	// If custom response format is specified, use templating
	if cmd.ResponseFormat != "" {
		// Convert formData to interface{} map for templating
		inputs := make(map[string]interface{})
		for k, v := range formData {
			inputs[k] = v
		}

		templateData := &TemplateData{
			Inputs:          inputs,
			WebhookResponse: webhookResp,
		}

		return b.Templater.RenderWithFallback(cmd.ResponseFormat, templateData)
	}

	// Fallback to original response format if no template specified
	response := fmt.Sprintf("✅ **Form Successfully Submitted**\n\n📋 **Command**: %s\n\n", strings.Title(cmd.Name))

	response += "**📝 Submitted Data:**\n"
	requiredFields := 0
	filledFields := 0

	for _, field := range cmd.Fields {
		if field.Required {
			requiredFields++
		}

		if value, exists := formData[field.Name]; exists && strings.TrimSpace(value) != "" {
			filledFields++

			icon := "📝"
			switch field.Type {
			case "attachment":
				icon = "📎"
			case "select":
				icon = "📋"
			case "remote_select":
				icon = "🌐"
			case "text":
				if strings.Contains(strings.ToLower(field.Name), "email") {
					icon = "📧"
				} else if strings.Contains(strings.ToLower(field.Name), "amount") || strings.Contains(strings.ToLower(field.Name), "price") || strings.Contains(strings.ToLower(field.Name), "cost") {
					icon = "💰"
				} else if strings.Contains(strings.ToLower(field.Name), "phone") {
					icon = "📞"
				} else if strings.Contains(strings.ToLower(field.Name), "url") || strings.Contains(strings.ToLower(field.Name), "link") {
					icon = "🔗"
				}
			}

			response += fmt.Sprintf("%s **%s**: %s\n", icon, strings.Title(field.Name), value)
		} else if field.Required {
			response += fmt.Sprintf("❌ **%s**: *Not provided*\n", strings.Title(field.Name))
		}
	}

	response += fmt.Sprintf("\n📊 **Summary**: %d/%d fields filled", filledFields, len(cmd.Fields))
	if requiredFields > 0 {
		response += fmt.Sprintf(" (%d required)", requiredFields)
	}

	if cmd.Webhook != "" {
		if webhookResp != nil && webhookResp.Status != "success" {
			response += fmt.Sprintf("\n\n❌ **Webhook Status**: Failed to send data\n🌐 **Endpoint**: %s\n⚠️ **Error**: %s", cmd.Webhook, webhookResp.Error)
		} else {
			response += fmt.Sprintf("\n\n✅ **Webhook Status**: Data sent successfully\n🌐 **Endpoint**: %s", cmd.Webhook)
		}
	}

	response += "\n\n✨ **Thank you for your submission!**"

	return response
}

func (b *Bot) createModalComponents(cmd *config.CommandSpec) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	for _, field := range cmd.Fields {
		if field.Type == "text" {
			style := discordgo.TextInputShort
			maxLength := 1000

			if strings.Contains(strings.ToLower(field.Name), "description") ||
				strings.Contains(strings.ToLower(field.Name), "details") ||
				strings.Contains(strings.ToLower(field.Name), "comment") {
				style = discordgo.TextInputParagraph
				maxLength = 4000
			}

			row := discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    field.Name,
						Label:       strings.Title(field.Name),
						Style:       style,
						Placeholder: fmt.Sprintf("Enter %s", field.Name),
						Required:    field.Required,
						MaxLength:   maxLength,
					},
				},
			}
			rows = append(rows, row)
		} else {
			log.Printf("Warning: Field type '%s' for field '%s' is not supported in Discord modals. Only 'text' fields are supported in modals.", field.Type, field.Name)
		}
	}

	return rows
}

func (b *Bot) validateFormData(cmd *config.CommandSpec, formData map[string]string) error {
	var errors []string

	for _, field := range cmd.Fields {
		value, exists := formData[field.Name]
		isEmpty := !exists || strings.TrimSpace(value) == ""

		switch field.Type {
		case "text":
			if field.Required && isEmpty {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			} else if exists && !isEmpty {
				if err := b.validateTextFormat(field, value); err != nil {
					errors = append(errors, fmt.Sprintf("• **%s** %s", strings.Title(field.Name), err.Error()))
				}
			}
		case "select":
			if field.Required && isEmpty {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			} else if exists && !isEmpty && len(field.Options) > 0 {
				valid := false
				for _, option := range field.Options {
					if value == option {
						valid = true
						break
					}
				}
				if !valid {
					availableOptions := strings.Join(field.Options, ", ")
					errors = append(errors, fmt.Sprintf("• **%s** has invalid value '%s'. Available options: %s", strings.Title(field.Name), value, availableOptions))
				}
			}
		case "remote_select":
			if field.Required && isEmpty {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			} else if exists && !isEmpty && field.Webhook != "" {
				remoteOptions, err := b.fetchRemoteOptions(field.Webhook)
				if err != nil {
					log.Printf("Failed to fetch remote options for validation: %v", err)
					errors = append(errors, fmt.Sprintf("• **%s** could not validate options (remote service unavailable)", strings.Title(field.Name)))
				} else {
					valid := false
					for _, option := range remoteOptions {
						if value == option.Value {
							valid = true
							break
						}
					}
					if !valid {
						errors = append(errors, fmt.Sprintf("• **%s** has invalid value '%s'", strings.Title(field.Name), value))
					}
				}
			}
		case "attachment":
			if field.Required && isEmpty {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Please fix the following issues:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

func (b *Bot) validateTextFormat(field config.FieldSpec, value string) error {
	fieldNameLower := strings.ToLower(field.Name)

	if strings.Contains(fieldNameLower, "email") {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("must be a valid email address")
		}
	}

	if strings.Contains(fieldNameLower, "amount") || strings.Contains(fieldNameLower, "price") || strings.Contains(fieldNameLower, "cost") {
		value = strings.ReplaceAll(value, ",", ".")
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("must be a valid number")
		}
	}

	if strings.Contains(fieldNameLower, "phone") {
		phoneRegex := regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$`)
		if !phoneRegex.MatchString(strings.ReplaceAll(value, " ", "")) {
			return fmt.Errorf("must be a valid phone number")
		}
	}

	if strings.Contains(fieldNameLower, "url") || strings.Contains(fieldNameLower, "link") {
		urlRegex := regexp.MustCompile(`^https?://[^\s]+$`)
		if !urlRegex.MatchString(value) {
			return fmt.Errorf("must be a valid URL starting with http:// or https://")
		}
	}

	return nil
}

func (b *Bot) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Printf("Error responding with error message: %v", err)
	}
}
