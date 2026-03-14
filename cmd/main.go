// Package main implements the Telegram bot entrypoint.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"bot-downloader/internal/config"
	"bot-downloader/internal/handlers"
	"bot-downloader/internal/logging"
	"bot-downloader/internal/services/ffmpeg"
	"bot-downloader/internal/services/ffprobe"
	"bot-downloader/internal/telegram"
	"bot-downloader/internal/telegram/gotelegrambot"
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

	ffmpegService, err := ffmpeg.New(cfg.FFmpegBinaryPath)
	if err != nil {
		logger.Error("failed to initialize ffmpeg service", "error", err)
		os.Exit(1)
	}
	metadataService, err := ffprobe.New(cfg.FFprobeBinaryPath)
	if err != nil {
		logger.Error("failed to initialize ffprobe service", "error", err)
		os.Exit(1)
	}

	tgClient, err := gotelegrambot.New(cfg.Token, logger, ffmpegService, metadataService)
	if err != nil {
		logger.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}

	username, err := tgClient.Username(context.Background())
	if err != nil {
		logger.Error("failed to fetch telegram bot profile", "error", err)
		os.Exit(1)
	}
	logger.Info("bot authorized", "username", username)

	var urlHandlers []handlers.Handler

	urlHandlers = append(urlHandlers, handlers.NewTiktokHandler(cfg.YtDlpBinaryPath, ffmpegService, logger))
	urlHandlers = append(urlHandlers, handlers.NewInstagramHandler(cfg.YtDlpBinaryPath, cfg.InstagramCookiesFilePath, ffmpegService, logger))

	if err := tgClient.Start(context.Background(), func(ctx context.Context, msg telegram.IncomingMessage) error {
		return routeMessage(ctx, tgClient, msg, cfg, urlHandlers, logger)
	}); err != nil {
		logger.Error("telegram client stopped with error", "error", err)
		os.Exit(1)
	}
}

func routeMessage(ctx context.Context, tg telegram.Client, msg telegram.IncomingMessage, cfg config.Config, handlers []handlers.Handler, logger *slog.Logger) error {
	if msg.ChatType == telegram.ChatTypePrivate {
		if _, ok := cfg.AllowedUserIDs[msg.UserID]; !ok {
			return nil
		}
		return handleMessage(ctx, tg, msg, handlers, logger)
	}

	if msg.ChatType == telegram.ChatTypeGroup || msg.ChatType == telegram.ChatTypeSupergroup {
		if _, ok := cfg.AllowedChatIDs[msg.ChatID]; !ok {
			return nil
		}
		return handleMessage(ctx, tg, msg, handlers, logger)
	}

	return nil
}

func handleMessage(ctx context.Context, tg telegram.Client, msg telegram.IncomingMessage, handlers []handlers.Handler, logger *slog.Logger) error {
	u, err := url.Parse(msg.Text)
	if err != nil {
		// this is not URL message, ignore it
		return nil
	}

	for _, h := range handlers {
		if h.Matcher(u) {
			logger.Info("matched message", "chat_id", msg.ChatID, "message_id", msg.MessageID, "url", u.String())

			if err := tg.DeleteMessage(ctx, msg.ChatID, msg.MessageID); err != nil {
				logger.Error(
					"failed to delete matched message",
					"chat_id",
					msg.ChatID,
					"message_id",
					msg.MessageID,
					"error",
					err,
				)
			}

			if err := h.Handle(ctx, tg, u, msg.ChatID); err != nil {
				return fmt.Errorf("handle matched url: %w", err)
			}

			return nil
		}
	}

	// don't found any handlers for given URL
	return nil
}
