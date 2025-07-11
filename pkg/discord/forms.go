package discord

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
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
		response += fmt.Sprintf("\n\n🔄 **Status**: Data will be sent to webhook\n🌐 **Endpoint**: %s", cmd.Webhook)
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
