// Package main implements the Telegram bot entrypoint.
package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"bot-downloader/internal/config"
	"bot-downloader/internal/handlers"
	"bot-downloader/internal/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := logging.New(cfg.AppEnv)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	logger.Info("bot authorized", "username", bot.Self.UserName)

	var urlHandlers []handlers.Handler

	urlHandlers = append(urlHandlers, handlers.NewTiktokHandler(cfg.YtDlpBinaryPath, logger))

	for update := range updates {
		if err := routeUpdate(bot, update, cfg, urlHandlers, logger); err != nil {
			logger.Error("failed to handle update", "error", err)
		}
	}
}

func routeUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg config.Config, handlers []handlers.Handler, logger *slog.Logger) error {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return nil
	}

	msg := update.Message

	if msg.Chat.IsPrivate() {
		if _, ok := cfg.AllowedUserIDs[msg.From.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, handlers, logger)
	}

	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if _, ok := cfg.AllowedChatIDs[msg.Chat.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, handlers, logger)
	}

	return nil
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, handlers []handlers.Handler, logger *slog.Logger) error {
	u, err := url.Parse(msg.Text)
	if err != nil {
		// this is not URL message, ignore it
		return nil
	}

	for _, h := range handlers {
		if h.Matcher(u) {
			logger.Info("matched message", "chat_id", msg.Chat.ID, "message_id", msg.MessageID, "url", u.String())

			deleteCfg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
			if _, err := bot.Request(deleteCfg); err != nil {
				logger.Error(
					"failed to delete matched message",
					"chat_id",
					msg.Chat.ID,
					"message_id",
					msg.MessageID,
					"error",
					err,
				)
			}

			if err := h.Handle(bot, u, msg.Chat.ID); err != nil {
				return fmt.Errorf("handle matched url: %w", err)
			}

			return nil
		}
	}

	// don't found any handlers for given URL
	return nil
}
