package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/pkg/postgres"
)

type SongRepo struct {
	*postgres.Postgres
}

func NewSongRepo(pg *postgres.Postgres) *SongRepo {
	return &SongRepo{pg}
}

func (r *SongRepo) Insert(ctx context.Context, song entity.Song) (int, error) {
	sql, args, _ := r.Builder.
		Insert("songs").
		Columns("song_name, group_name, link, release_date").
		Values(song.Name, song.Group, song.Link, song.ReleaseDate).
		Suffix("RETURNING id").
		ToSql()

	var id int
	err := r.GetQueryRunner(ctx).QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return 0, ErrAlreadyExists
			}
		}
		return 0, fmt.Errorf("SongRepo.Insert - QueryRow: %w", err)
	}

	return id, nil
}

func (r *SongRepo) GetById(ctx context.Context, songId int) (entity.Song, error) {
	sql, args, _ := r.Builder.
		Select("id, song_name, group_name, link, release_date").
		From("songs").
		Where("id = ?", songId).
		ToSql()

	var song entity.Song
	err := r.GetQueryRunner(ctx).QueryRow(ctx, sql, args...).Scan(
		&song.Id,
		&song.Name,
		&song.Group,
		&song.Link,
		&song.ReleaseDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Song{}, ErrNotFound
		}
		return entity.Song{}, fmt.Errorf("SongRepo.GetById - QueryRow: %w", err)
	}

	return song, nil
}

func (r *SongRepo) UpdateById(ctx context.Context, songId int, song entity.Song) error {
	return nil
}

func (r *SongRepo) DeleteById(ctx context.Context, songId int) error {
	sql, args, _ := r.Builder.Delete("songs").
		Where("id = ?", songId).
		ToSql()

	_, err := r.GetQueryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SongRepo.DeleteById - Exec: %w", err)
	}

	return nil
}
