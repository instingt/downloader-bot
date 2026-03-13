// Package gotelegrambot implements telegram.Client using github.com/go-telegram/bot.
package gotelegrambot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"bot-downloader/internal/telegram"
)

const pollTimeout = 31 * time.Second

type Client struct {
	bot    *telegrambot.Bot
	logger *slog.Logger

	mu             sync.RWMutex
	messageHandler telegram.MessageHandler
}

func New(token string, logger *slog.Logger) (*Client, error) {
	client := &Client{
		logger: logger,
	}

	b, err := telegrambot.New(
		token,
		telegrambot.WithHTTPClient(
			pollTimeout,
			&http.Client{Timeout: pollTimeout},
		),
		telegrambot.WithErrorsHandler(client.handleBotError),
		telegrambot.WithDefaultHandler(client.handleUpdate),
		telegrambot.WithWorkers(1),
	)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot client: %w", err)
	}

	client.bot = b
	return client, nil
}

func (c *Client) Start(ctx context.Context, handler telegram.MessageHandler) error {
	if handler == nil {
		return errors.New("message handler is nil")
	}

	c.mu.Lock()
	c.messageHandler = handler
	c.mu.Unlock()

	c.bot.Start(ctx)
	return nil
}

func (c *Client) Username(ctx context.Context) (string, error) {
	user, err := c.bot.GetMe(ctx)
	if err != nil {
		return "", fmt.Errorf("get bot profile: %w", err)
	}

	return user.Username, nil
}

func (c *Client) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	ok, err := c.bot.DeleteMessage(ctx, &telegrambot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: messageID,
	})
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	if !ok {
		return errors.New("delete message: telegram api returned false")
	}

	return nil
}

func (c *Client) SendVideoFile(ctx context.Context, chatID int64, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open video file: %w", err)
	}

	_, err = c.bot.SendVideo(ctx, &telegrambot.SendVideoParams{
		ChatID: chatID,
		Video: &models.InputFileUpload{
			Filename: filepath.Base(path),
			Data:     file,
		},
	})
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("send video: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close video file: %w", closeErr)
	}

	return nil
}

func (c *Client) SendDocumentFile(ctx context.Context, chatID int64, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open document file: %w", err)
	}

	_, err = c.bot.SendDocument(ctx, &telegrambot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: filepath.Base(path),
			Data:     file,
		},
	})
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("send document: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close document file: %w", closeErr)
	}

	return nil
}

func (c *Client) handleBotError(err error) {
	if c.logger == nil {
		return
	}
	c.logger.Error("telegram polling error", "error", err)
}

func (c *Client) handleUpdate(ctx context.Context, _ *telegrambot.Bot, update *models.Update) {
	if update == nil || update.Message == nil || update.Message.From == nil {
		return
	}

	msg := update.Message
	incoming := telegram.IncomingMessage{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		UserID:    msg.From.ID,
		ChatType:  string(msg.Chat.Type),
		Text:      msg.Text,
	}

	c.mu.RLock()
	handler := c.messageHandler
	c.mu.RUnlock()
	if handler == nil {
		return
	}

	if err := handler(ctx, incoming); err != nil && c.logger != nil {
		c.logger.Error("failed to handle incoming message", "error", err)
	}
}
