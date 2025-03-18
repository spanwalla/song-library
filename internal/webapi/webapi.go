package webapi

import (
	"context"
	"time"
)

//go:generate mockgen -source=webapi.go -destination=../mocks/webapi/mock.go -package=webapimocks

type GetSongInfoOutput struct {
	ReleaseDate time.Time
	Text        string
	Link        string
}

type SongInfo interface {
	Get(ctx context.Context, group, song string) (GetSongInfoOutput, error)
}
