// Package video provides functionality for work with video files.
package video

import "context"

type Transcoder interface {
	Transcode(ctx context.Context, path string) (string, error)
}
