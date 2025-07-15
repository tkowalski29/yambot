package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"
)

// TemplateData holds all data available for template rendering
type TemplateData struct {
	Inputs          map[string]interface{} `json:"inputs"`
	WebhookResponse *WebhookResponse       `json:"webhook_response,omitempty"`
}

// WebhookResponse holds the response data from webhook calls
type WebhookResponse struct {
	StatusCode int                    `json:"status_code"`
	Status     string                 `json:"status"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// ResponseTemplater handles response template rendering
type ResponseTemplater struct{}

// NewResponseTemplater creates a new ResponseTemplater instance
func NewResponseTemplater() *ResponseTemplater {
	return &ResponseTemplater{}
}

// RenderTemplate renders a template with the provided data
func (rt *ResponseTemplater) RenderTemplate(templateStr string, data *TemplateData) (string, error) {
	if strings.TrimSpace(templateStr) == "" {
		return "✅ Komenda została przyjęta.", nil
	}

	// Create template with custom functions
	tmpl, err := template.New("response").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"json": func(v interface{}) string {
			bytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return string(bytes)
		},
	}).Parse(templateStr)

	if err != nil {
		log.Printf("Template parse error: %v", err)
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// RenderWithFallback renders a template with fallback to default message on error
func (rt *ResponseTemplater) RenderWithFallback(templateStr string, data *TemplateData) string {
	result, err := rt.RenderTemplate(templateStr, data)
	if err != nil {
		log.Printf("Template rendering failed, using fallback: %v", err)
		return "✅ Komenda została przyjęta."
	}
	return result
}

// ValidateTemplate validates a template string without rendering it
func (rt *ResponseTemplater) ValidateTemplate(templateStr string) error {
	if strings.TrimSpace(templateStr) == "" {
		return nil
	}

	// Create template with same custom functions as in RenderTemplate
	_, err := template.New("validation").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"json": func(v interface{}) string {
			return fmt.Sprintf("%v", v) // Simplified for validation
		},
	}).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("template validation error: %w", err)
	}

	return nil
}