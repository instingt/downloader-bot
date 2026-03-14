package video

import "context"

type VideoMetadata struct {
	Width    int
	Height   int
	Duration int
}

type MetadataReader interface {
	ReadMetadata(ctx context.Context, videoPath string) (VideoMetadata, error)
}
