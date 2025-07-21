# GitHub Actions Workflows

## Docker Build Workflow

### Release Workflow (`docker-release.yml`)

**Trigger:** Automatycznie uruchamia się gdy wydasz nowy release na GitHub

**Co robi:**
- Buduje obraz Docker z tagami semver (np. `v1.0.0`, `1.0`, `1`)
- Pushuje obraz do GitHub Container Registry (ghcr.io)
- Używa cache dla szybszych buildów

**Tagowanie:**
- `latest` - dla głównej gałęzi
- `release` - dla release'ów
- `v1.0.0` - pełna wersja
- `1.0` - major.minor
- `1` - tylko major

## Jak używać

### Wydanie nowego release:

1. Przejdź do sekcji "Releases" na GitHub
2. Kliknij "Create a new release"
3. Ustaw tag (np. `v1.0.0`)
4. Napisz opis release
5. Kliknij "Publish release"

Workflow automatycznie zbuduje i opublikuje obraz Docker.

### Sprawdzenie obrazów:

Obrazy będą dostępne w GitHub Container Registry:
```
ghcr.io/{username}/yambot:latest
ghcr.io/{username}/yambot:v1.0.0
ghcr.io/{username}/yambot:release
```

### Uruchomienie lokalnie:

```bash
docker run -d \
  -e DISCORD_TOKEN=your_token \
  -v /path/to/config.yml:/app/config/config.yml \
  ghcr.io/{username}/yambot:latest
```

## Wymagania

- Repository musi mieć włączone GitHub Packages
- `GITHUB_TOKEN` jest automatycznie dostępny
- Dockerfile musi być w głównym katalogu 