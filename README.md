# yambot

A Discord bot that handles slash commands and modal forms with webhook integration. The bot automatically sends form data and file attachments to configured webhook endpoints.

## Configuration

The bot is configured using a `config.yml` file with the following structure:

### Basic Structure

```yaml
bot:
  discord:
    token: DISCORD_TOKEN

commands:
  - name: command-name
    type: slash|modal
    webhook: "https://webhook-url"
    fields:
      - name: field-name
        type: field-type
        required: true|false
        options: ["option1", "option2"]  # for select fields
        webhook: "https://webhook-url"   # for remote_select fields
```

### Field Types

#### text
Simple text input field.

```yaml
- name: description
  type: text
  required: true
```

#### textarea
Multi-line text input field for longer content (descriptions, comments, feedback).

```yaml
- name: message
  type: textarea
  required: true
```

#### select
Dropdown selection with predefined options.

```yaml
- name: priority
  type: select
  options: ["High", "Medium", "Low"]
  required: true
```

#### remote_select
Dropdown selection with options loaded from a remote webhook.

```yaml
- name: department
  type: remote_select
  webhook: "https://api.example.com/departments"
  required: true
```

#### attachment
File upload field. **Note**: Attachment fields are only supported in slash commands, not in modal forms.

```yaml
- name: document
  type: attachment
  required: true
```

### Command Types

#### Slash Commands
Commands that appear in Discord's slash command interface. All field types are supported, including file attachments.

```yaml
- name: simple-command
  type: slash
  webhook: "https://webhook-url/submit"
  fields:
    - name: message
      type: text
      required: true
```

#### Modal Forms
Pop-up forms for more complex input. **Note**: Only `text` and `textarea` field types are supported in modal forms due to Discord's limitations.

```yaml
- name: feedback-form
  type: modal
  webhook: "https://webhook-url/feedback"
  fields:
    - name: name
      type: text
      required: true
    - name: email
      type: text
      required: true
    - name: rating
      type: select
      options: ["1", "2", "3", "4", "5"]
      required: true
    - name: comments
      type: text
      required: false
```

### Complete Examples

#### Slash Command with Remote Select

```yaml
commands:
  - name: expense-report
    type: slash
    webhook: "https://n8n.local/webhooks/submit-expense"
    fields:
      - name: company
        type: remote_select
        webhook: "https://api.company.com/departments"
        required: true
      - name: amount
        type: text
        required: true
      - name: receipt
        type: attachment
        required: true
      - name: category
        type: select
        options: ["Travel", "Office", "Marketing", "Other"]
        required: true
```

#### Modal Form with Mixed Field Types

```yaml
commands:
  - name: support-ticket
    type: modal
    webhook: "https://support.company.com/webhook"
    fields:
      - name: title
        type: text
        required: true
      - name: priority
        type: select
        options: ["Low", "Medium", "High", "Critical"]
        required: true
      - name: department
        type: remote_select
        webhook: "https://api.company.com/departments"
        required: true
      - name: description
        type: textarea
        required: true
      - name: attachment
        type: attachment
        required: false
```

#### Feedback Form with Textarea

```yaml
commands:
  - name: feedback
    type: modal
    webhook: "https://webhook-url/feedback"
    fields:
      - name: subject
        type: text
        required: true
      - name: message
        type: textarea
        required: true
      - name: rating
        type: select
        options: ["1", "2", "3", "4", "5"]
        required: true
```

### Field Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Field identifier and label |
| `type` | string | Yes | Field type (text, textarea, select, remote_select, attachment) |
| `required` | boolean | Yes | Whether the field is mandatory |
| `options` | array | No | Available options for select fields |
| `webhook` | string | No | Webhook URL for remote_select fields |

### Command Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Command name (appears in Discord) |
| `type` | string | Yes | Command type (slash or modal) |
| `webhook` | string | Yes | Webhook URL to send form data |
| `fields` | array | Yes | Array of field definitions |

## Webhook Integration

### Data Format

The bot sends data to webhooks in JSON format with the following structure:

#### For Slash Commands
```json
{
  "command": "command-name",
  "field1": "value1",
  "field2": "value2",
  "attachment_field": "File: filename.pdf (application/pdf, 1024000 bytes)",
  "attachment_field_url": "https://cdn.discordapp.com/attachments/...",
  "attachment_field_content_type": "application/pdf",
  "attachment_field_size": "1024000"
}
```

#### For Modal Forms
```json
{
  "field1": "value1",
  "field2": "value2"
}
```

### Attachment Handling

For slash commands with file attachments, the bot sends:
- **File information**: Filename, content type, and size
- **Download URL**: Direct link to the file on Discord's CDN
- **Metadata**: Content type and file size for processing

### Webhook Response

The bot expects webhooks to return HTTP status codes:
- **200-299**: Success (data processed successfully)
- **Other codes**: Error (will be reported to the user)

### Error Handling

If a webhook fails to receive data, the bot will:
1. Log the error for debugging
2. Display an error message to the user
3. Continue normal operation for other commands

## Architecture

### WebhookService

The bot uses a dedicated `WebhookService` for handling all webhook operations:

- **Centralized webhook management**: All webhook operations are handled by a single service
- **Support for both command types**: Handles both slash commands and modal forms
- **File attachment support**: Properly processes and sends file metadata for attachments
- **Error handling**: Comprehensive error handling and logging
- **Timeout management**: Configurable timeouts for webhook requests

### File Structure

```
yambot/
├── cmd/
│   └── main.go              # Application entry point
├── pkg/
│   ├── config/
│   │   ├── config.go        # Configuration management
│   │   └── config_test.go   # Configuration tests
│   └── discord/
│       ├── bot.go           # Main bot logic
│       ├── commands.go      # Command registration
│       ├── forms.go         # Modal form handling
│       ├── webhook.go       # Webhook service
│       └── forms_test.go    # Form handling tests
├── config.yml               # Configuration file
├── go.mod                   # Go module file
└── README.md               # This documentation
```

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Discord Bot Token
- Webhook endpoints for receiving data

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd yambot
```

2. Install dependencies:
```bash
go mod download
```

3. Create a configuration file `config.yml` (see Configuration section above)

4. Set your Discord bot token:
```bash
export DISCORD_TOKEN="your-discord-bot-token"
```

5. Build and run the bot:
```bash
go build -o yambot ./cmd
./yambot
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DISCORD_TOKEN` | Discord bot token | Yes |

### Testing

Run the test suite:
```bash
go test ./...
```

### Docker

Build and run with Docker:
```bash
docker build -t yambot .
docker run -e DISCORD_TOKEN="your-token" -v $(pwd)/config.yml:/app/config.yml yambot
```