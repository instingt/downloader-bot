# Repository Guidelines

## Project Structure & Module Organization
- Entry point: `cmd/main.go` (Telegram bot startup and update routing).
- Internal packages: `internal/config/config.go` (environment-based config loading and validation).
- Build outputs: `bin/` (for compiled binaries such as `bot-linux-amd64`).
- Dependency and tool config: `go.mod`, `go.sum`, `.golangci.yml`, `Makefile`.
- Environment values are read from shell variables and optionally `.env`.

Keep reusable logic inside `internal/` packages and keep `cmd/` focused on wiring and process lifecycle.

## Build, Test, and Development Commands
- `make run`: validates required env vars, loads `.env` when present, and runs the bot.
- `make test`: runs all Go tests (`go test ./...`).
- `make lint`: runs `golangci-lint` with configured linters.
- `make tidy`: normalizes module dependencies (`go mod tidy`).
- `make build-linux-amd64`: builds a static Linux amd64 binary into `bin/`.

Example:
```bash
TELEGRAM_BOT_TOKEN=... ALLOWED_TELEGRAM_USER_IDS=123 ALLOWED_TELEGRAM_CHAT_IDS=-1001 make run
```

## Coding Style & Naming Conventions
- Follow standard Go formatting (`gofmt`) and idiomatic Go style.
- Use tabs for indentation (Go default).
- Package names: short, lowercase, no underscores.
- Exported identifiers: `CamelCase`; unexported: `camelCase`.
- Keep functions focused; return wrapped errors with context where helpful.
- Ensure package comments exist (enforced by `revive`).

## Testing Guidelines
- Place tests next to code as `*_test.go` files.
- Prefer table-driven tests for parsing/validation logic (for example, config env parsing).
- Run `make test` locally before opening a PR.
- Add regression tests for bug fixes and edge cases (empty env values, invalid IDs, duplicates).

## Commit & Pull Request Guidelines
- Commit style in history uses Conventional Commits, e.g. `chore: add linter`.
- Recommended prefixes: `feat:`, `fix:`, `chore:`, `refactor:`, `test:`.
- Keep commits small and single-purpose.
- PRs should include a clear behavior summary, linked issue (if applicable), test/lint evidence (`make test`, `make lint`), and explicit notes for any env var changes.

## Security & Configuration Tips
- Never commit `.env` or secrets.
- Treat `TELEGRAM_BOT_TOKEN` as sensitive.
- Restrict `ALLOWED_TELEGRAM_USER_IDS` and `ALLOWED_TELEGRAM_CHAT_IDS` to known IDs only.
