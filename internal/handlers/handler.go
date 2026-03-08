// Package handlers contains message handlers for provider-specific URLs.
package handlers

import (
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	Matcher(*url.URL) bool
	Handle(*tgbotapi.BotAPI, *url.URL, int64) error
}
