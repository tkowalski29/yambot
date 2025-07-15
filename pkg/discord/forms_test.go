package discord

import (
	"strings"
	"testing"

	"yambot/pkg/config"
)

func TestValidateFormData_RequiredFields(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
			{Name: "amount", Type: "text", Required: true},
			{Name: "optional", Type: "text", Required: false},
		},
	}

	tests := []struct {
		name      string
		formData  map[string]string
		shouldErr bool
	}{
		{
			name:      "valid data",
			formData:  map[string]string{"title": "Test Title", "amount": "100"},
			shouldErr: false,
		},
		{
			name:      "missing required title",
			formData:  map[string]string{"amount": "100"},
			shouldErr: true,
		},
		{
			name:      "empty required title",
			formData:  map[string]string{"title": "", "amount": "100"},
			shouldErr: true,
		},
		{
			name:      "whitespace only title",
			formData:  map[string]string{"title": "   ", "amount": "100"},
			shouldErr: true,
		},
		{
			name:      "missing optional field",
			formData:  map[string]string{"title": "Test Title", "amount": "100"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.validateFormData(cmd, tt.formData)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateFormData() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateFormData_SelectFields(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}, Required: true},
			{Name: "optional_select", Type: "select", Options: []string{"Option 1", "Option 2"}, Required: false},
		},
	}

	tests := []struct {
		name      string
		formData  map[string]string
		shouldErr bool
	}{
		{
			name:      "valid option",
			formData:  map[string]string{"company": "Company A"},
			shouldErr: false,
		},
		{
			name:      "invalid option",
			formData:  map[string]string{"company": "Company C"},
			shouldErr: true,
		},
		{
			name:      "empty required selection",
			formData:  map[string]string{"company": ""},
			shouldErr: true,
		},
		{
			name:      "empty optional selection",
			formData:  map[string]string{"company": "Company A", "optional_select": ""},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.validateFormData(cmd, tt.formData)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateFormData() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateFormData_AttachmentFields(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "pdf", Type: "attachment", Required: true},
			{Name: "optional_file", Type: "attachment", Required: false},
		},
	}

	tests := []struct {
		name      string
		formData  map[string]string
		shouldErr bool
	}{
		{
			name:      "valid attachment",
			formData:  map[string]string{"pdf": "file.pdf"},
			shouldErr: false,
		},
		{
			name:      "missing required attachment",
			formData:  map[string]string{},
			shouldErr: true,
		},
		{
			name:      "empty required attachment",
			formData:  map[string]string{"pdf": ""},
			shouldErr: true,
		},
		{
			name:      "missing optional attachment",
			formData:  map[string]string{"pdf": "file.pdf"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.validateFormData(cmd, tt.formData)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateFormData() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateFormData_MultipleErrors(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}, Required: true},
			{Name: "pdf", Type: "attachment", Required: true},
		},
	}

	formData := map[string]string{
		"title":   "",
		"company": "Invalid Company",
		"pdf":     "",
	}

	err := bot.validateFormData(cmd, formData)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "Title") {
		t.Error("Expected error message to contain 'Title'")
	}
	if !strings.Contains(errorMsg, "Company") {
		t.Error("Expected error message to contain 'Company'")
	}
	if !strings.Contains(errorMsg, "Pdf") {
		t.Error("Expected error message to contain 'Pdf'")
	}
}

func TestCreateFormResponse(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name:    "cost",
		Webhook: "https://example.com/webhook",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
			{Name: "amount", Type: "text", Required: true},
		},
	}

	formData := map[string]string{
		"title":  "Test Title",
		"amount": "100",
	}

	response := bot.createFormResponse(cmd, formData, nil)

	if !strings.Contains(response, "Form Successfully Submitted") {
		t.Error("Expected response to contain success message")
	}
	if !strings.Contains(response, "Cost") {
		t.Error("Expected response to contain command name")
	}
	if !strings.Contains(response, "Test Title") {
		t.Error("Expected response to contain form data")
	}
	if !strings.Contains(response, "100") {
		t.Error("Expected response to contain form data")
	}
	if !strings.Contains(response, "https://example.com/webhook") {
		t.Error("Expected response to contain webhook URL")
	}
	if !strings.Contains(response, "Data sent successfully") {
		t.Error("Expected response to contain success status")
	}
}

func TestCreateFormResponse_WebhookError(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name:    "cost",
		Webhook: "https://example.com/webhook",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
		},
	}

	formData := map[string]string{
		"title": "Test Title",
	}

	webhookResp := &WebhookResponse{
		Status: "error",
		Error:  "connection failed",
	}
	response := bot.createFormResponse(cmd, formData, webhookResp)

	if !strings.Contains(response, "Form Successfully Submitted") {
		t.Error("Expected response to contain success message")
	}
	if !strings.Contains(response, "Failed to send data") {
		t.Error("Expected response to contain webhook error status")
	}
	if !strings.Contains(response, "connection failed") {
		t.Error("Expected response to contain webhook error message")
	}
}

func TestCreateModalComponents(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
			{Name: "description", Type: "text", Required: false},
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}, Required: true},
		},
	}

	components := bot.createModalComponents(cmd)

	if len(components) != 2 {
		t.Errorf("Expected 2 components (only text fields), got %d", len(components))
	}
}

func TestValidateTextFormat(t *testing.T) {
	bot := &Bot{}

	tests := []struct {
		name        string
		field       config.FieldSpec
		value       string
		shouldErr   bool
		description string
	}{
		{
			name:        "valid email",
			field:       config.FieldSpec{Name: "email", Type: "text"},
			value:       "test@example.com",
			shouldErr:   false,
			description: "Valid email format",
		},
		{
			name:        "invalid email",
			field:       config.FieldSpec{Name: "email", Type: "text"},
			value:       "invalid-email",
			shouldErr:   true,
			description: "Invalid email format",
		},
		{
			name:        "valid amount",
			field:       config.FieldSpec{Name: "amount", Type: "text"},
			value:       "123.45",
			shouldErr:   false,
			description: "Valid amount format",
		},
		{
			name:        "invalid amount",
			field:       config.FieldSpec{Name: "amount", Type: "text"},
			value:       "invalid-amount",
			shouldErr:   true,
			description: "Invalid amount format",
		},
		{
			name:        "valid phone",
			field:       config.FieldSpec{Name: "phone", Type: "text"},
			value:       "+1234567890",
			shouldErr:   false,
			description: "Valid phone format",
		},
		{
			name:        "invalid phone",
			field:       config.FieldSpec{Name: "phone", Type: "text"},
			value:       "invalid-phone",
			shouldErr:   true,
			description: "Invalid phone format",
		},
		{
			name:        "valid url",
			field:       config.FieldSpec{Name: "url", Type: "text"},
			value:       "https://example.com",
			shouldErr:   false,
			description: "Valid URL format",
		},
		{
			name:        "invalid url",
			field:       config.FieldSpec{Name: "url", Type: "text"},
			value:       "invalid-url",
			shouldErr:   true,
			description: "Invalid URL format",
		},
		{
			name:        "non-special field",
			field:       config.FieldSpec{Name: "title", Type: "text"},
			value:       "Any value",
			shouldErr:   false,
			description: "Non-special field should pass any value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bot.validateTextFormat(tt.field, tt.value)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateTextFormat() error = %v, shouldErr %v, test: %s", err, tt.shouldErr, tt.description)
			}
		})
	}
}

func TestValidateFormData_RequiredFalse(t *testing.T) {
	bot := &Bot{}

	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text", Required: true},
			{Name: "description", Type: "text", Required: false},
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}, Required: false},
			{Name: "optional_file", Type: "attachment", Required: false},
		},
	}

	formData := map[string]string{
		"title": "Required Title",
	}

	err := bot.validateFormData(cmd, formData)
	if err != nil {
		t.Errorf("Expected no error for optional fields, got: %v", err)
	}
}
