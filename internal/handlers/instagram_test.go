package handlers

import (
	"net/url"
	"testing"
)

func TestInstagramHandlerMatcher(t *testing.T) {
	t.Parallel()

	h := NewInstagramHandler("/usr/bin/yt-dlp", "/tmp/cookies.txt", nil)

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "reel with www host",
			raw:  "https://www.instagram.com/reel/C8abc123xyz/",
			want: true,
		},
		{
			name: "reel with root host",
			raw:  "https://instagram.com/reel/C8abc123xyz",
			want: true,
		},
		{
			name: "reel with subdomain host",
			raw:  "https://m.instagram.com/reel/C8abc123xyz/",
			want: true,
		},
		{
			name: "non https",
			raw:  "http://www.instagram.com/reel/C8abc123xyz/",
			want: false,
		},
		{
			name: "non instagram host",
			raw:  "https://example.com/reel/C8abc123xyz/",
			want: false,
		},
		{
			name: "post path not reel",
			raw:  "https://www.instagram.com/p/C8abc123xyz/",
			want: false,
		},
		{
			name: "tv path not reel",
			raw:  "https://www.instagram.com/tv/C8abc123xyz/",
			want: false,
		},
		{
			name: "profile url",
			raw:  "https://www.instagram.com/someuser/",
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tc.raw)
			if err != nil {
				t.Fatalf("parse url: %v", err)
			}

			got := h.Matcher(u)
			if got != tc.want {
				t.Fatalf("expected %v, got %v for URL %q", tc.want, got, tc.raw)
			}
		})
	}
}
