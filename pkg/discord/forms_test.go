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
			{Name: "title", Type: "text"},
			{Name: "amount", Type: "text"},
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
			name:      "missing title",
			formData:  map[string]string{"amount": "100"},
			shouldErr: true,
		},
		{
			name:      "empty title",
			formData:  map[string]string{"title": "", "amount": "100"},
			shouldErr: true,
		},
		{
			name:      "whitespace only title",
			formData:  map[string]string{"title": "   ", "amount": "100"},
			shouldErr: true,
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
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}},
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
			name:      "empty selection",
			formData:  map[string]string{"company": ""},
			shouldErr: true,
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
			{Name: "pdf", Type: "attachment"},
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
			name:      "missing attachment",
			formData:  map[string]string{},
			shouldErr: true,
		},
		{
			name:      "empty attachment",
			formData:  map[string]string{"pdf": ""},
			shouldErr: true,
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
			{Name: "title", Type: "text"},
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}},
			{Name: "pdf", Type: "attachment"},
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
		Name: "cost",
		Webhook: "https://example.com/webhook",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text"},
			{Name: "amount", Type: "text"},
		},
	}

	formData := map[string]string{
		"title":  "Test Title",
		"amount": "100",
	}

	response := bot.createFormResponse(cmd, formData)

	if !strings.Contains(response, "Form Successfully Submitted") {
		t.Error("Expected response to contain success message")
	}
	if !strings.Contains(response, "cost") {
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
}

func TestCreateModalComponents(t *testing.T) {
	bot := &Bot{}
	
	cmd := &config.CommandSpec{
		Name: "test",
		Fields: []config.FieldSpec{
			{Name: "title", Type: "text"},
			{Name: "description", Type: "text"},
			{Name: "company", Type: "select", Options: []string{"Company A", "Company B"}},
		},
	}

	components := bot.createModalComponents(cmd)

	if len(components) != 2 {
		t.Errorf("Expected 2 components (only text fields), got %d", len(components))
	}
}