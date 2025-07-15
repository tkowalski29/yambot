# yambot

A Discord bot that handles slash commands and modal forms with webhook integration.

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
File upload field.

```yaml
- name: document
  type: attachment
  required: true
```

### Command Types

#### Slash Commands
Commands that appear in Discord's slash command interface.

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
Pop-up forms for more complex input.

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