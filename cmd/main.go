// Package main implements the Telegram bot entrypoint.
package main

import (
	"log"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"bot-downloader/internal/config"
	"bot-downloader/internal/handlers"
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

	var urlHandlers []handlers.Handler

	urlHandlers = append(urlHandlers, handlers.NewTiktokHandler())

	for update := range updates {
		if err := routeUpdate(bot, update, cfg, urlHandlers); err != nil {
			log.Printf("failed to handle update: %v", err)
		}
	}
}

func routeUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg config.Config, handlers []handlers.Handler) error {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return nil
	}

	msg := update.Message

	if msg.Chat.IsPrivate() {
		if _, ok := cfg.AllowedUserIDs[msg.From.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, handlers)
	}

	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if _, ok := cfg.AllowedChatIDs[msg.Chat.ID]; !ok {
			return nil
		}
		return handleMessage(bot, msg, handlers)
	}

	return nil
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, handlers []handlers.Handler) error {
	u, err := url.Parse(msg.Text)
	if err != nil {
		// this is not URL message, ignore it
		return nil
	}

	for _, h := range handlers {
		if h.Matcher(u) {
			return h.Handle(bot, u)
		}
	}

	// don't found any handlers for given URL
	return nil
}
