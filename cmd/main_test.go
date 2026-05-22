// Package main tests the bot entrypoint helpers.
package main

import "testing"

func TestExtractSingleURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
		ok   bool
	}{
		{
			name: "plain url",
			text: "https://www.tiktok.com/@user/video/123",
			want: "https://www.tiktok.com/@user/video/123",
			ok:   true,
		},
		{
			name: "url after words",
			text: "Look a video I found in tiktok https://www.tiktok.com/@user/video/123",
			want: "https://www.tiktok.com/@user/video/123",
			ok:   true,
		},
		{
			name: "url before words",
			text: "https://www.tiktok.com/@user/video/123 this is interesting",
			want: "https://www.tiktok.com/@user/video/123",
			ok:   true,
		},
		{
			name: "url with trailing punctuation",
			text: "Check this: https://www.tiktok.com/@user/video/123.",
			want: "https://www.tiktok.com/@user/video/123",
			ok:   true,
		},
		{
			name: "url adjacent to text",
			text: "Check this:https://www.tiktok.com/@user/video/123",
			want: "https://www.tiktok.com/@user/video/123",
			ok:   true,
		},
		{
			name: "no url",
			text: "Look a video I found in tiktok",
			ok:   false,
		},
		{
			name: "invalid url",
			text: "Look at https://",
			ok:   false,
		},
		{
			name: "multiple urls",
			text: "https://www.tiktok.com/@user/video/123 https://www.instagram.com/reel/abc/",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := extractSingleURL(tt.text)
			if ok != tt.ok {
				t.Fatalf("extractSingleURL() ok = %v, want %v", ok, tt.ok)
			}
			if !ok {
				return
			}
			if got.String() != tt.want {
				t.Fatalf("extractSingleURL() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}
