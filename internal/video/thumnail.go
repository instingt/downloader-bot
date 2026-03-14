package video

import "context"

type Thumbnailer interface {
	CreateThumnail(ctx context.Context, path string) (string, error)
}
