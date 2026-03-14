// Package telegram defines library-agnostic Telegram client contracts.
package telegram

import "context"

const (
	ChatTypePrivate    = "private"
	ChatTypeGroup      = "group"
	ChatTypeSupergroup = "supergroup"
)

type IncomingMessage struct {
	ChatID    int64
	MessageID int
	UserID    int64
	ChatType  string
	Text      string
}

type MessageHandler func(context.Context, IncomingMessage) error

type Client interface {
	Start(context.Context, MessageHandler) error
	Username(context.Context) (string, error)
	DeleteMessage(context.Context, int64, int) error
	SendVideoFile(context.Context, int64, string) error
	SendDocumentFile(context.Context, int64, string) error
	SendVideoWithDocumentFallback(context.Context, int64, string) error
}
