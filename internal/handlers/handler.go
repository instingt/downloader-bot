// Package handlers contains message handlers for provider-specific URLs.
package handlers

import (
	"context"
	"net/url"

	"bot-downloader/internal/telegram"
)

type Handler interface {
	Matcher(*url.URL) bool
	Handle(context.Context, telegram.Client, *url.URL, int64) error
}
