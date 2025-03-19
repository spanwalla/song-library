package repository

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/pkg/postgres"
)

type CoupletRepo struct {
	*postgres.Postgres
}

func NewCoupletRepo(pg *postgres.Postgres) *CoupletRepo {
	return &CoupletRepo{pg}
}

func (r *CoupletRepo) Insert(ctx context.Context, couplets []entity.Couplet) error {
	query := r.Builder.
		Insert("couplets").
		Columns("song_id", "sequence_number", "couplet_text")

	for _, couplet := range couplets {
		query = query.Values(couplet.SongId, couplet.SequenceNumber, couplet.Text)
	}

	sql, args, _ := query.ToSql()
	log.Debugf("SongRepo.Insert - ToSql: %s", sql)

	_, err := r.GetQueryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SongRepo.Insert - Exec: %w", err)
	}

	return nil
}

func (r *CoupletRepo) GetBySongId(ctx context.Context, songId, offset, limit int) ([]entity.Couplet, error) {
	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	} else if limit <= 0 {
		limit = defaultPaginationLimit
	}

	if offset < 0 {
		offset = 0
	}

	sql, args, _ := r.Builder.
		Select("song_id, sequence_number, couplet_text").
		From("couplets").
		Where("song_id = ?", songId).
		OrderBy("sequence_number").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()

	cmdTag, err := r.GetQueryRunner(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("CoupletRepo.GetBySongId - Query: %w", err)
	}
	defer cmdTag.Close()

	couplets := make([]entity.Couplet, 0)
	for cmdTag.Next() {
		var couplet entity.Couplet
		err = cmdTag.Scan(&couplet.SongId, &couplet.SequenceNumber, &couplet.Text)
		if err != nil {
			return nil, fmt.Errorf("CoupletRepo.GetBySongId - Scan: %w", err)
		}
		couplets = append(couplets, couplet)
	}

	return couplets, nil
}

func (r *CoupletRepo) GetAvailableSequenceNumber(ctx context.Context, songId int) (int, error) {
	sql, args, _ := r.Builder.
		Select("COALESCE(MAX(sequence_number), 0) + 1").
		From("couplets").
		Where("song_id = ?", songId).
		ToSql()

	var sequenceNumber int
	err := r.GetQueryRunner(ctx).QueryRow(ctx, sql, args...).Scan(&sequenceNumber)
	if err != nil {
		return 0, fmt.Errorf("CoupletRepo.GetAvailableSequenceNumber - QueryRow: %w", err)
	}

	return sequenceNumber, nil
}

func (r *CoupletRepo) GetCoupletsCount(ctx context.Context, songId int) (int, error) {
	sql, args, _ := r.Builder.
		Select("COUNT(*)").
		From("couplets").
		Where("song_id = ?", songId).
		ToSql()

	var count int
	err := r.GetQueryRunner(ctx).QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CoupletRepo.GetCoupletsCount - QueryRow: %w", err)
	}

	return count, nil
}

func (r *CoupletRepo) DeleteBySongId(ctx context.Context, songId int) error {
	sql, args, _ := r.Builder.
		Delete("couplets").
		Where("song_id = ?", songId).
		ToSql()

	_, err := r.GetQueryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("CoupletRepo.DeleteBySongId - Exec: %w", err)
	}

	return nil
}
