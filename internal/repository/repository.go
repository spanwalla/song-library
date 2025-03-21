package repository

import (
	"context"
	"time"

	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/pkg/postgres"
)

//go:generate mockgen -source=repository.go -destination=../mocks/repository/mock.go -package=repomocks

const (
	maxPaginationLimit     = 10
	defaultPaginationLimit = 5
)

// Transactor определяет интерфейс для работы с транзакциями.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UpdateSongInput struct {
	Name        *string
	Group       *string
	Link        *string
	ReleaseDate *time.Time
}

type Song interface {
	Insert(ctx context.Context, song entity.Song) (int, error)
	GetById(ctx context.Context, songId int) (entity.Song, error)
	Search(ctx context.Context, filters map[string]string, orderBy [][]string, offset, limit int) ([]entity.Song, error)
	UpdateById(ctx context.Context, songId int, input UpdateSongInput) error
	DeleteById(ctx context.Context, songId int) error
}

type Couplet interface {
	Insert(ctx context.Context, couplets []entity.Couplet) error
	GetBySongId(ctx context.Context, songId, offset, limit int) ([]entity.Couplet, error)
	GetAvailableSequenceNumber(ctx context.Context, songId int) (int, error)
	GetCoupletsCount(ctx context.Context, songId int) (int, error)
	DeleteBySongId(ctx context.Context, songId int) error
}

type Repositories struct {
	Song
	Couplet
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		Song:    NewSongRepo(pg),
		Couplet: NewCoupletRepo(pg),
	}
}
