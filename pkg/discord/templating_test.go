package discord

import (
	"testing"
)

func TestResponseTemplater_RenderTemplate(t *testing.T) {
	templater := NewResponseTemplater()

	tests := []struct {
		name     string
		template string
		data     *TemplateData
		expected string
		hasError bool
	}{
		{
			name:     "empty template returns default message",
			template: "",
			data:     &TemplateData{},
			expected: "✅ Komenda została przyjęta.",
			hasError: false,
		},
		{
			name:     "simple input templating",
			template: "Hello {{ .Inputs.name }}!",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "John",
				},
			},
			expected: "Hello John!",
			hasError: false,
		},
		{
			name:     "webhook response templating",
			template: "Status: {{ .WebhookResponse.Status }}",
			data: &TemplateData{
				WebhookResponse: &WebhookResponse{
					Status: "success",
				},
			},
			expected: "Status: success",
			hasError: false,
		},
		{
			name:     "nested data templating",
			template: "Ticket ID: {{ .WebhookResponse.Data.ticket_id }}",
			data: &TemplateData{
				WebhookResponse: &WebhookResponse{
					Status: "success",
					Data: map[string]interface{}{
						"ticket_id": "12345",
					},
				},
			},
			expected: "Ticket ID: 12345",
			hasError: false,
		},
		{
			name:     "function upper",
			template: "Name: {{ upper .Inputs.name }}",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "john doe",
				},
			},
			expected: "Name: JOHN DOE",
			hasError: false,
		},
		{
			name:     "function lower",
			template: "Name: {{ lower .Inputs.name }}",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "JOHN DOE",
				},
			},
			expected: "Name: john doe",
			hasError: false,
		},
		{
			name:     "function title",
			template: "Name: {{ title .Inputs.name }}",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "john doe",
				},
			},
			expected: "Name: John Doe",
			hasError: false,
		},
		{
			name:     "complex template with multiple variables",
			template: "✅ Zgłoszono koszt **{{ .Inputs.title }}** na kwotę **{{ .Inputs.amount }} PLN**\n📎 Plik: {{ .Inputs.attachment.name }}\n🔗 Status: {{ .WebhookResponse.Status }}",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"title":  "Zakup materiałów",
					"amount": "150.00",
					"attachment": map[string]interface{}{
						"name": "receipt.pdf",
					},
				},
				WebhookResponse: &WebhookResponse{
					Status: "success",
				},
			},
			expected: "✅ Zgłoszono koszt **Zakup materiałów** na kwotę **150.00 PLN**\n📎 Plik: receipt.pdf\n🔗 Status: success",
			hasError: false,
		},
		{
			name:     "invalid template syntax",
			template: "Hello {{ .Inputs.name",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "John",
				},
			},
			expected: "",
			hasError: true,
		},
		{
			name:     "template with missing data",
			template: "Hello {{ .Inputs.missing }}!",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "John",
				},
			},
			expected: "Hello <no value>!",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := templater.RenderTemplate(tt.template, tt.data)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResponseTemplater_RenderWithFallback(t *testing.T) {
	templater := NewResponseTemplater()

	tests := []struct {
		name     string
		template string
		data     *TemplateData
		expected string
	}{
		{
			name:     "valid template",
			template: "Hello {{ .Inputs.name }}!",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "John",
				},
			},
			expected: "Hello John!",
		},
		{
			name:     "invalid template falls back to default",
			template: "Hello {{ .Inputs.name",
			data: &TemplateData{
				Inputs: map[string]interface{}{
					"name": "John",
				},
			},
			expected: "✅ Komenda została przyjęta.",
		},
		{
			name:     "empty template returns default",
			template: "",
			data:     &TemplateData{},
			expected: "✅ Komenda została przyjęta.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := templater.RenderWithFallback(tt.template, tt.data)
			
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResponseTemplater_ValidateTemplate(t *testing.T) {
	templater := NewResponseTemplater()

	tests := []struct {
		name     string
		template string
		hasError bool
	}{
		{
			name:     "valid template",
			template: "Hello {{ .Inputs.name }}!",
			hasError: false,
		},
		{
			name:     "empty template is valid",
			template: "",
			hasError: false,
		},
		{
			name:     "template with functions",
			template: "{{ upper .Inputs.name }}",
			hasError: false,
		},
		{
			name:     "template with conditions",
			template: "{{ if .Inputs.name }}Hello {{ .Inputs.name }}{{ end }}",
			hasError: false,
		},
		{
			name:     "invalid template syntax - missing closing brace",
			template: "Hello {{ .Inputs.name",
			hasError: true,
		},
		{
			name:     "invalid template syntax - malformed condition",
			template: "{{ if .Inputs.name Hello {{ end }}",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := templater.ValidateTemplate(tt.template)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestWebhookResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *WebhookResponse
		template string
		expected string
	}{
		{
			name: "success response",
			response: &WebhookResponse{
				StatusCode: 200,
				Status:     "success",
				Data: map[string]interface{}{
					"id":         "12345",
					"created_at": "2023-01-01T10:00:00Z",
				},
			},
			template: "ID: {{ .WebhookResponse.Data.id }}, Status: {{ .WebhookResponse.Status }}",
			expected: "ID: 12345, Status: success",
		},
		{
			name: "error response",
			response: &WebhookResponse{
				StatusCode: 500,
				Status:     "error",
				Error:      "Internal server error",
			},
			template: "Error: {{ .WebhookResponse.Error }} ({{ .WebhookResponse.StatusCode }})",
			expected: "Error: Internal server error (500)",
		},
		{
			name: "no webhook response",
			response: nil,
			template: "Status: {{ if .WebhookResponse }}{{ .WebhookResponse.Status }}{{ else }}No webhook{{ end }}",
			expected: "Status: No webhook",
		},
	}

	templater := NewResponseTemplater()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &TemplateData{
				WebhookResponse: tt.response,
			}

			result, err := templater.RenderTemplate(tt.template, data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}