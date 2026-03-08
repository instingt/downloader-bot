// Package main implements the Telegram bot entrypoint.
package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"bot-downloader/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatalf("failed to initialize telegram bot: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	log.Printf("bot authorized as %s", bot.Self.UserName)

	for update := range updates {
		if err := routeUpdate(bot, update, cfg); err != nil {
			log.Printf("failed to handle update: %v", err)
		}
	}
}

func routeUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg config.Config) error {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return nil
	}

	msg := update.Message

	if msg.Chat.IsPrivate() {
		if _, ok := cfg.AllowedUserIDs[msg.From.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, "dm")
	}

	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if _, ok := cfg.AllowedChatIDs[msg.Chat.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, "group")
	}

	return nil
}

func handleMessage(_ *tgbotapi.BotAPI, msg *tgbotapi.Message, scope string) error {
	textLen := len(strings.TrimSpace(msg.Text))
	log.Printf(
		"accepted message scope=%s chat_id=%d user_id=%d message_id=%d text_len=%d",
		scope,
		msg.Chat.ID,
		msg.From.ID,
		msg.MessageID,
		textLen,
	)
	return nil
}
