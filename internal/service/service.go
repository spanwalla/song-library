package service

import (
	"context"
	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/internal/repository"
	"github.com/spanwalla/song-library/internal/webapi"
	"time"
)

//go:generate mockgen -source=service.go -destination=../mocks/service/mock.go -package=servicemocks

type InsertSongInput struct {
	Group string
	Song  string
}

type UpdateSongInput struct {
	Name        *string
	Group       *string
	Link        *string
	ReleaseDate *time.Time
	Text        *string
}

type GetTextInput struct {
	SongId int
	Offset int
	Limit  int
}

type Song interface {
	Insert(ctx context.Context, input InsertSongInput) error
	// Get TODO: Фильтрация по всем полям и пагинация
	Get(ctx context.Context, songId int) (entity.Song, error)
	GetText(ctx context.Context, input GetTextInput) ([]string, int, error)
	Update(ctx context.Context, songId int, input UpdateSongInput) error
	Delete(ctx context.Context, songId int) error
}

type Services struct {
	Song
}

type Dependencies struct {
	Repos      *repository.Repositories
	SongInfo   webapi.SongInfo
	Transactor repository.Transactor
}

func NewServices(deps Dependencies) *Services {
	return &Services{
		Song: NewSongService(deps.Repos.Song, deps.Repos.Couplet, deps.Transactor, deps.SongInfo),
	}
}
