# bot-downloader

`bot-downloader` is a Go-based Telegram bot that watches incoming messages for supported social media links, downloads the video, and sends it back into Telegram.

Currently the bot supports:

- TikTok links on `https://*.tiktok.com`
- Instagram Reel links on `https://*.instagram.com/reel/...`

The bot is access-restricted. It only responds to configured Telegram user IDs in private chats and configured chat IDs in group or supergroup chats.

## How It Works

1. The bot starts and loads its configuration from environment variables.
2. It validates required binaries and files such as `yt-dlp`, `ffmpeg`, `ffprobe`, and the Instagram cookies file.
3. When a supported URL is posted in an allowed chat, the bot:
   - detects the matching handler
   - downloads the video with `yt-dlp`
   - transcodes it with `ffmpeg`
   - sends the result back to Telegram

## Requirements

Before running the project, make sure you have:

- Go installed
- `yt-dlp` installed and available in `PATH`
- `ffmpeg` installed and available in `PATH`
- `ffprobe` installed and available in `PATH`
- a Telegram bot token from BotFather
- allowed Telegram user IDs and/or chat IDs
- an Instagram cookies file for Reel downloads

## Configuration

The application reads values from shell environment variables and can also load them from a local `.env` file.

Required variables:

- `TELEGRAM_BOT_TOKEN`: Telegram bot token
- `ALLOWED_TELEGRAM_USER_IDS`: comma-separated list of allowed private user IDs
- `ALLOWED_TELEGRAM_CHAT_IDS`: comma-separated list of allowed group or supergroup chat IDs
- `INSTAGRAM_COOKIES_FILE_PATH`: path to a valid cookies file used for Instagram downloads

Optional variables:

- `APP_ENV`: `development` or `production` (defaults to `development`)

Example `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:example-token
ALLOWED_TELEGRAM_USER_IDS=123456789,987654321
ALLOWED_TELEGRAM_CHAT_IDS=-1001234567890
INSTAGRAM_COOKIES_FILE_PATH=./instagram.cookies
APP_ENV=development
```

## Run The Project

### 1. Install dependencies

Make sure the required binaries are installed and available in your shell:

```bash
go version
yt-dlp --version
ffmpeg -version
ffprobe -version
```

### 2. Prepare configuration

Create a `.env` file in the project root or export the variables in your shell.

### 3. Start the bot

Run:

```bash
make run
```

The `run` target:

- loads `.env` if it exists
- validates the required environment variables
- starts the bot with `go run ./cmd/main.go`

You can also run it by passing variables directly:

```bash
TELEGRAM_BOT_TOKEN=... \
ALLOWED_TELEGRAM_USER_IDS=123456789 \
ALLOWED_TELEGRAM_CHAT_IDS=-1001234567890 \
INSTAGRAM_COOKIES_FILE_PATH=./instagram.cookies \
make run
```

## Development Commands

- `make run`: validate config and run the bot locally
- `make test`: run all Go tests
- `make lint`: run `golangci-lint`
- `make tidy`: clean up `go.mod` and `go.sum`
- `make build-linux-amd64`: build a Linux amd64 binary into `bin/`
- `make reload-instagram-cookies`: recreate `instagram.cookies` using browser cookies via `yt-dlp`

## Build

To build a Linux amd64 binary:

```bash
make build-linux-amd64
```

Output:

```text
bin/bot-linux-amd64
```

## Notes

- Do not commit `.env` or any secrets.
- Treat `TELEGRAM_BOT_TOKEN` as sensitive.
- Restrict allowed Telegram IDs to trusted users and chats only.
- Instagram downloads depend on a valid cookies file, so refresh it when it expires.
