package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"yambot/pkg/config"
)

func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	response := b.createFormResponse(commandSpec, formData)

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

func (b *Bot) createFormResponse(cmd *config.CommandSpec, formData map[string]string) string {
	response := fmt.Sprintf("✅ **Form Successfully Submitted**\n\n📋 **Command**: %s\n\n", cmd.Name)

	response += "**📝 Submitted Data:**\n"
	for _, field := range cmd.Fields {
		if value, exists := formData[field.Name]; exists && strings.TrimSpace(value) != "" {
			response += fmt.Sprintf("• **%s**: %s\n", strings.Title(field.Name), value)
		}
	}

	if cmd.Webhook != "" {
		response += fmt.Sprintf("\n🔄 **Status**: Data will be sent to webhook: %s", cmd.Webhook)
	}

	response += "\n\n✨ Thank you for your submission!"

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
						Required:    true,
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
		switch field.Type {
		case "text":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			} else if len(strings.TrimSpace(value)) < 1 {
				errors = append(errors, fmt.Sprintf("• **%s** cannot be empty", strings.Title(field.Name)))
			}
		case "select":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			} else if len(field.Options) > 0 {
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
		case "attachment":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				errors = append(errors, fmt.Sprintf("• **%s** is required", strings.Title(field.Name)))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Please fix the following issues:\n%s", strings.Join(errors, "\n"))
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
