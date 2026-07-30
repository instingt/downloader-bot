package handlers

import (
	"net/url"
	"testing"
)

func TestYouTubeShortsHandlerMatcher(t *testing.T) {
	t.Parallel()

	handler := NewYouTubeShortsHandler("", nil, nil)
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "www youtube shorts",
			raw:  "https://www.youtube.com/shorts/abc123",
			want: true,
		},
		{
			name: "youtube shorts",
			raw:  "https://youtube.com/shorts/abc123",
			want: true,
		},
		{
			name: "mobile youtube shorts with query parameters",
			raw:  "https://m.youtube.com/shorts/abc123?feature=share",
			want: true,
		},
		{
			name: "shorts path with trailing slash",
			raw:  "https://www.youtube.com/shorts/abc123/",
			want: true,
		},
		{
			name: "http scheme",
			raw:  "http://www.youtube.com/shorts/abc123",
			want: false,
		},
		{
			name: "youtu be short link",
			raw:  "https://youtu.be/abc123",
			want: false,
		},
		{
			name: "regular youtube video",
			raw:  "https://www.youtube.com/watch?v=abc123",
			want: false,
		},
		{
			name: "shorts path without slash",
			raw:  "https://www.youtube.com/shorts",
			want: false,
		},
		{
			name: "unsupported youtube subdomain",
			raw:  "https://music.youtube.com/shorts/abc123",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tt.raw)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}

			if got := handler.Matcher(u); got != tt.want {
				t.Errorf("Matcher(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
