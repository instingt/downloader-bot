# YouTube Shorts Downloads

Make it possible to download YouTube Shorts using the same workflow as TikTok downloads.

## Supported URLs

Accept HTTPS YouTube Shorts URLs with the `/shorts/<id>` path on these hosts:

- `www.youtube.com`
- `youtube.com`
- `m.youtube.com`

URLs may include query parameters, which must be passed to the downloader unchanged.

Do not accept `youtu.be` links or other YouTube URL paths, even if they redirect to a Short.

The matcher does not need additional validation of the Shorts ID or special handling for a trailing slash.

## Download Workflow

For a matched YouTube Shorts URL, use the existing TikTok workflow:

1. Delete the matched Telegram message.
2. Download the video with the configured `yt-dlp` binary.
3. Transcode the downloaded video with the configured video transcoder.
4. Send the result to the source chat as a video, falling back to a document when required.
5. Remove temporary downloaded and transcoded files after processing completes.

Implement this in a dedicated YouTube Shorts URL handler and register it with the existing message router. Other routing, authorization, download, transcoding, upload, cleanup, logging, and error behavior must remain the same as the TikTok handler.

If processing fails, log the detailed error server-side and use the existing generic handled-error Telegram response.

## Acceptance Criteria

- A message containing exactly one supported HTTPS `/shorts/` URL is routed to the YouTube Shorts handler.
- URLs on `www.youtube.com`, `youtube.com`, and `m.youtube.com` are supported.
- Query parameters are preserved when invoking `yt-dlp`.
- Non-Shorts YouTube URLs and `youtu.be` URLs are ignored.
- Successful downloads follow the existing TikTok download, transcode, upload, and cleanup flow.
- Download or processing failures use the existing generic handled-error response.
- Tests cover the handler's accepted and rejected URL matching behavior.
