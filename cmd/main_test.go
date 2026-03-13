package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"testing"

	"bot-downloader/internal/config"
	"bot-downloader/internal/handlers"
	"bot-downloader/internal/telegram"
)

type fakeTelegramClient struct {
	deleteCalls int
	deleteErr   error
}

func (f *fakeTelegramClient) Start(context.Context, telegram.MessageHandler) error {
	return nil
}

func (f *fakeTelegramClient) Username(context.Context) (string, error) {
	return "bot", nil
}

func (f *fakeTelegramClient) DeleteMessage(context.Context, int64, int) error {
	f.deleteCalls++
	return f.deleteErr
}

func (f *fakeTelegramClient) SendVideoFile(context.Context, int64, string) error {
	return nil
}

func (f *fakeTelegramClient) SendDocumentFile(context.Context, int64, string) error {
	return nil
}

type fakeHandler struct {
	matcher func(*url.URL) bool
	called  int
}

func (h *fakeHandler) Matcher(u *url.URL) bool {
	if h.matcher == nil {
		return false
	}
	return h.matcher(u)
}

func (h *fakeHandler) Handle(context.Context, telegram.Client, *url.URL, int64) error {
	h.called++
	return nil
}

func TestRouteMessageAllowlist(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		AllowedUserIDs: map[int64]struct{}{
			42: {},
		},
		AllowedChatIDs: map[int64]struct{}{
			-1001: {},
		},
	}

	tests := []struct {
		name          string
		msg           telegram.IncomingMessage
		wantHandleCnt int
	}{
		{
			name: "allowed private user",
			msg: telegram.IncomingMessage{
				UserID:    42,
				ChatID:    10,
				ChatType:  telegram.ChatTypePrivate,
				MessageID: 1,
				Text:      "https://www.tiktok.com/@user/video/1",
			},
			wantHandleCnt: 1,
		},
		{
			name: "disallowed private user",
			msg: telegram.IncomingMessage{
				UserID:    77,
				ChatID:    10,
				ChatType:  telegram.ChatTypePrivate,
				MessageID: 1,
				Text:      "https://www.tiktok.com/@user/video/1",
			},
			wantHandleCnt: 0,
		},
		{
			name: "allowed group",
			msg: telegram.IncomingMessage{
				UserID:    42,
				ChatID:    -1001,
				ChatType:  telegram.ChatTypeGroup,
				MessageID: 1,
				Text:      "https://www.tiktok.com/@user/video/1",
			},
			wantHandleCnt: 1,
		},
		{
			name: "disallowed supergroup",
			msg: telegram.IncomingMessage{
				UserID:    42,
				ChatID:    -1002,
				ChatType:  telegram.ChatTypeSupergroup,
				MessageID: 1,
				Text:      "https://www.tiktok.com/@user/video/1",
			},
			wantHandleCnt: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tg := &fakeTelegramClient{}
			h := &fakeHandler{
				matcher: func(*url.URL) bool { return true },
			}

			err := routeMessage(context.Background(), tg, tc.msg, cfg, []handlers.Handler{h}, logger)
			if err != nil {
				t.Fatalf("routeMessage returned error: %v", err)
			}

			if h.called != tc.wantHandleCnt {
				t.Fatalf("expected handler calls %d, got %d", tc.wantHandleCnt, h.called)
			}
		})
	}
}

func TestHandleMessageDeleteFailureStillProcessesHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tg := &fakeTelegramClient{deleteErr: errors.New("delete failed")}
	h := &fakeHandler{
		matcher: func(*url.URL) bool { return true },
	}

	msg := telegram.IncomingMessage{
		ChatID:    -1001,
		MessageID: 77,
		Text:      "https://www.tiktok.com/@user/video/1",
	}

	err := handleMessage(context.Background(), tg, msg, []handlers.Handler{h}, logger)
	if err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}

	if tg.deleteCalls != 1 {
		t.Fatalf("expected delete calls 1, got %d", tg.deleteCalls)
	}
	if h.called != 1 {
		t.Fatalf("expected handler call despite delete failure")
	}
}
