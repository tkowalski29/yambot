# Templating w YamBot

YamBot oferuje możliwość dostosowania format odpowiedzi dla komend Discord za pomocą systemu szablonowania opartego na Go templates.

## Podstawowe użycie

W konfiguracji YAML możesz dodać pole `response_format` do każdej komendy:

```yaml
commands:
  - name: cost
    type: slash
    webhook: "https://your-webhook.com/endpoint"
    response_format: |
      ✅ Zgłoszono koszt **{{ .Inputs.title }}** na kwotę **{{ .Inputs.amount }} PLN**
      📎 Plik: {{ .Inputs.pdf.name }}
      🔗 Status: {{ .WebhookResponse.Status }}
    fields:
      - name: title
        type: text
        required: true
      - name: amount  
        type: text
        required: true
      - name: pdf
        type: attachment
        required: true
```

## Dostępne zmienne

### inputs

Obiekt `inputs` zawiera wszystkie dane wprowadzone przez użytkownika:

- **Pola tekstowe**: `{{ .Inputs.field_name }}`
- **Pola select**: `{{ .Inputs.field_name }}`
- **Załączniki**: `{{ .Inputs.attachment_name.name }}`, `{{ .Inputs.attachment_name.size }}`, `{{ .Inputs.attachment_name.content_type }}`

#### Przykład dla attachment:
```yaml
response_format: |
  📎 **Załącznik**: {{ .Inputs.document.name }}
  📏 **Rozmiar**: {{ .Inputs.document.size }} bajtów
  🎭 **Typ**: {{ .Inputs.document.content_type }}
```

### webhook_response

Obiekt `webhook_response` zawiera odpowiedź z webhook:

- `{{ .WebhookResponse.Status }}` - status odpowiedzi ("success", "error", "200", itp.)
- `{{ .WebhookResponse.Status_code }}` - kod HTTP (200, 404, 500, itp.)
- `{{ .WebhookResponse.data }}` - dane JSON zwrócone przez webhook
- `{{ .WebhookResponse.Error }}` - opis błędu (jeśli wystąpił)

#### Przykład z zagnieżdżonymi danymi:
```yaml
response_format: |
  🆔 **ID zgłoszenia**: {{ .WebhookResponse.Data.ticket_id }}
  📅 **Data utworzenia**: {{ .WebhookResponse.Data.created_at }}
  👤 **Przypisano do**: {{ .WebhookResponse.Data.assignee.name }}
```

## Funkcje pomocnicze

### Formatowanie tekstu

```yaml
response_format: |
  🔤 **Wielkie litery**: {{ upper .Inputs.title }}
  🔡 **Małe litery**: {{ lower .Inputs.title }}
  🎩 **Tytułowe**: {{ title .Inputs.title }}
```

### Formatowanie JSON

```yaml
response_format: |
  📋 **Pełna odpowiedź**:
  ```json
  {{ json .WebhookResponse.data }}
  ```
```

## Warunki logiczne

```yaml
response_format: |
  ✅ **Status**: {{ if eq .WebhookResponse.Status "success" }}Pomyślnie zapisano{{ else }}Wystąpił błąd{{ end }}
  
  {{ if .WebhookResponse.Data.urgent }}
  🚨 **PILNE**: Zgłoszenie wymaga natychmiastowej uwagi!
  {{ end }}
```

## Pętle

```yaml
response_format: |
  📋 **Lista elementów**:
  {{ range .WebhookResponse.Data.items }}
  • {{ .name }} ({{ .Status }})
  {{ end }}
```

## Przykłady praktyczne

### 1. Zgłoszenie kosztów

```yaml
commands:
  - name: cost
    type: slash
    webhook: "https://api.example.com/costs"
    response_format: |
      💰 **Zgłoszenie kosztów przyjęte**
      
      **Szczegóły:**
      • **Tytuł**: {{ .Inputs.title }}
      • **Kwota**: {{ .Inputs.amount }} PLN
      • **Firma**: {{ .Inputs.company }}
      • **Miesiąc**: {{ .Inputs.month }}
      {{ if .Inputs.pdf }}• **Załącznik**: {{ .Inputs.pdf.name }} ({{ .Inputs.pdf.size }} B){{ end }}
      
      {{ if eq .WebhookResponse.Status "success" }}
      ✅ **Status**: Zapisano w systemie (ID: {{ .WebhookResponse.Data.id }})
      🔗 **Link**: {{ .WebhookResponse.Data.view_url }}
      {{ else }}
      ❌ **Błąd**: {{ .WebhookResponse.Error }}
      {{ end }}
    fields:
      - name: title
        type: text
        required: true
      - name: amount
        type: text
        required: true
      - name: company
        type: select
        options: ["Firma A", "Firma B"]
        required: true
      - name: month
        type: select
        options: ["Styczeń", "Luty", "Marzec"]
        required: true
      - name: pdf
        type: attachment
        required: false
```

### 2. Daily standup

```yaml
commands:
  - name: daily
    type: modal
    webhook: "https://api.example.com/standup"
    response_format: |
      📅 **Daily standup zapisany**
      
      **Co robiłeś wczoraj:**
      {{ .Inputs.yesterday }}
      
      **Co planujesz dzisiaj:**
      {{ .Inputs.today }}
      
      **Problemy/blokady:**
      {{ .Inputs.problems }}
      
      {{ if eq .WebhookResponse.Status "success" }}
      ✅ Zapisano w systemie Jira ({{ .WebhookResponse.Data.jira_ticket }})
      📊 **Statystyki tygodnia**: {{ .WebhookResponse.Data.week_stats.completed }}/{{ .WebhookResponse.Data.week_stats.planned }} zadań
      {{ else }}
      ❌ **Błąd zapisywania**: {{ .WebhookResponse.Error }}
      {{ end }}
    fields:
      - name: yesterday
        type: text
        required: true
      - name: today
        type: text
        required: true
      - name: problems
        type: text
        required: false
```

### 3. Zgłoszenie błędu

```yaml
commands:
  - name: bug_report
    type: modal
    webhook: "https://api.example.com/bugs"
    response_format: |
      🐞 **Dziękujemy za zgłoszenie błędu!**
      
      **Tytuł**: {{ .Inputs.title }}
      **Priorytet**: {{ upper .Inputs.priority }}
      **Opis**: {{ .Inputs.description }}
      
      {{ if eq .WebhookResponse.Status "success" }}
      ✅ **Utworzono ticket**: #{{ .WebhookResponse.Data.ticket_number }}
      👤 **Przypisano do**: {{ .WebhookResponse.Data.assignee }}
      ⏱️ **Szacowany czas naprawy**: {{ .WebhookResponse.Data.estimated_fix_time }}
      
      {{ if eq .Inputs.priority "high" }}
      🚨 **WYSOKI PRIORYTET** - zespół został powiadomiony przez Slack
      {{ end }}
      {{ else }}
      ❌ **Nie udało się utworzyć ticket**: {{ .WebhookResponse.Error }}
      📞 **Skontaktuj się z**: support@example.com
      {{ end }}
    fields:
      - name: title
        type: text
        required: true
      - name: description
        type: text
        required: true
      - name: priority
        type: select
        options: ["low", "medium", "high", "critical"]
        required: true
```

## Obsługa błędów

Jeśli szablon zawiera błędy składniowe lub błędy wykonania, bot automatycznie użyje domyślnej wiadomości: `"✅ Komenda została przyjęta."`

Błędy szablonów są logowane, więc możesz je debugować w logach aplikacji.

## Walidacja szablonów

Szablony są walidowane podczas ładowania konfiguracji. Jeśli szablon zawiera błędy składniowe, aplikacja nie wystartuje i pokaże komunikat o błędzie.

Przykład błędu:
```
template validation failed: invalid template in command 'cost': template: validation:1: unexpected "}" in operand
```

## Domyślne zachowanie

Jeśli nie określisz `response_format`, bot użyje domyślnej wiadomości odpowiedzi:
- **Slash commands**: `"✅ Komenda została przyjęta."`
- **Modal forms**: Szczegółowa odpowiedź z listą przesłanych danych (zachowanie poprzednie)