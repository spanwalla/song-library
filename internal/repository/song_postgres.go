package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	log "github.com/sirupsen/logrus"

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

func (r *SongRepo) Search(ctx context.Context, filters map[string]string, orderBy [][]string, offset, limit int) ([]entity.Song, error) {
	validColumnsMapping := map[string]string{
		"id":          "id",
		"group":       "group_name",
		"song":        "song_name",
		"link":        "link",
		"releaseDate": "release_date",
	}

	query := r.Builder.
		Select("id, group_name, song_name, link, release_date").
		From("songs")

	for field, value := range filters {
		if dbColumn, ok := validColumnsMapping[field]; ok {
			query = query.Where(fmt.Sprintf("%s = ?", dbColumn), value)
		}
	}

	for _, field := range orderBy {
		if dbColumn, ok := validColumnsMapping[field[0]]; ok {
			if strings.ToLower(field[1]) == "asc" || strings.ToLower(field[1]) == "desc" {
				query = query.OrderBy(fmt.Sprintf("%s %s", dbColumn, field[1]))
			}
		}
	}

	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	} else if limit <= 0 {
		limit = defaultPaginationLimit
	}

	if offset < 0 {
		offset = 0
	}

	sql, args, _ := query.Offset(uint64(offset)).Limit(uint64(limit)).ToSql()
	log.Debugf("SongRepo.Search - sql: %s", sql)

	cmdTag, err := r.GetQueryRunner(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SongRepo.Search - Query: %w", err)
	}
	defer cmdTag.Close()

	songs := make([]entity.Song, 0)
	for cmdTag.Next() {
		var song entity.Song
		err = cmdTag.Scan(&song.Id, &song.Group, &song.Name, &song.Link, &song.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("SongRepo.Search - Scan: %w", err)
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func buildUpdateMap(input UpdateSongInput) map[string]any {
	updates := make(map[string]any)
	if input.Name != nil {
		updates["song_name"] = *input.Name
	}
	if input.Group != nil {
		updates["group_name"] = *input.Group
	}
	if input.Link != nil {
		updates["link"] = *input.Link
	}
	if input.ReleaseDate != nil {
		updates["release_date"] = *input.ReleaseDate
	}
	return updates
}

func (r *SongRepo) UpdateById(ctx context.Context, songId int, input UpdateSongInput) error {
	updates := buildUpdateMap(input)
	if len(updates) == 0 {
		return nil
	}

	sql, args, _ := r.Builder.
		Update("songs").
		Where("id = ?", songId).
		SetMap(updates).
		ToSql()

	_, err := r.GetQueryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SongRepo.UpdateById - Exec: %w", err)
	}

	return nil
}

func (r *SongRepo) DeleteById(ctx context.Context, songId int) error {
	sql, args, _ := r.Builder.
		Delete("songs").
		Where("id = ?", songId).
		ToSql()

	_, err := r.GetQueryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SongRepo.DeleteById - Exec: %w", err)
	}

	return nil
}
