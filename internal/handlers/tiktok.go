package handlers

import (
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TikTokHandler struct {
	ytDlpBinary string
}

func NewTiktokHandler(ytDlpBinary string) *TikTokHandler {
	return &TikTokHandler{
		ytDlpBinary: ytDlpBinary,
	}
}

func (h *TikTokHandler) Matcher(u *url.URL) bool {
	return false
}

func (h *TikTokHandler) Handle(bot *tgbotapi.BotAPI, u *url.URL) error {
	return nil
}
