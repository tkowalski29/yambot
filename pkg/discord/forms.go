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
	response := fmt.Sprintf("Form submitted for command: %s\n\n", cmd.Name)

	for _, field := range cmd.Fields {
		if value, exists := formData[field.Name]; exists {
			response += fmt.Sprintf("**%s**: %s\n", field.Name, value)
		}
	}

	if cmd.Webhook != "" {
		response += fmt.Sprintf("\nData will be sent to webhook: %s", cmd.Webhook)
	}

	return response
}

func (b *Bot) createModalComponents(cmd *config.CommandSpec) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent

	for _, field := range cmd.Fields {
		if field.Type == "text" {
			row := discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    field.Name,
						Label:       strings.Title(field.Name),
						Style:       discordgo.TextInputShort,
						Placeholder: fmt.Sprintf("Enter %s", field.Name),
						Required:    true,
						MaxLength:   1000,
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
	for _, field := range cmd.Fields {
		switch field.Type {
		case "text":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				return fmt.Errorf("field %s is required", field.Name)
			}
		case "select":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				return fmt.Errorf("field %s is required", field.Name)
			} else if len(field.Options) > 0 {
				valid := false
				for _, option := range field.Options {
					if value == option {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("field %s has invalid value: %s", field.Name, value)
				}
			}
		case "attachment":
			if value, exists := formData[field.Name]; !exists || strings.TrimSpace(value) == "" {
				return fmt.Errorf("field %s is required", field.Name)
			}
		}
	}

	return nil
}
