package service

import (
	"context"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/internal/repository"
	"github.com/spanwalla/song-library/internal/webapi"
)

type SongService struct {
	songRepo    repository.Song
	coupletRepo repository.Couplet
	transactor  repository.Transactor
	songInfo    webapi.SongInfo
}

func NewSongService(songRepo repository.Song, coupletRepo repository.Couplet, transactor repository.Transactor, songInfo webapi.SongInfo) *SongService {
	return &SongService{
		songRepo:    songRepo,
		coupletRepo: coupletRepo,
		transactor:  transactor,
		songInfo:    songInfo,
	}
}

func (s *SongService) Insert(ctx context.Context, input InsertSongInput) error {
	info, err := s.songInfo.Get(ctx, input.Group, input.Song)
	if err != nil {
		log.Errorf("SongService.Insert - s.songInfo.Get: %v", err)
		return ErrCannotGetSongInfo
	}

	songId, err := s.songRepo.Insert(ctx, entity.Song{
		Name:        input.Song,
		Group:       input.Group,
		Link:        info.Link,
		ReleaseDate: info.ReleaseDate,
	})
	if err != nil {
		log.Errorf("SongService.Insert - s.songRepo.Insert: %v", err)
		return ErrCannotInsertSong
	}

	var couplets []entity.Couplet
	splitText := strings.Split(info.Text, "\n\n")
	for i, piece := range splitText {
		couplets = append(couplets, entity.Couplet{
			SongId:         songId,
			SequenceNumber: i + 1,
			Text:           piece,
		})
	}

	err = s.coupletRepo.Insert(ctx, couplets)
	if err != nil {
		log.Errorf("SongService.Insert - s.coupletRepo.Insert: %v", err)
		return ErrCannotInsertCouplets
	}

	return nil
}

func (s *SongService) Get(ctx context.Context, songId int) (entity.Song, error) {
	song, err := s.songRepo.GetById(ctx, songId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Song{}, ErrSongNotFound
		}
		log.Errorf("SongService.Get - s.songRepo.GetById: %v", err)
		return entity.Song{}, ErrCannotGetSong
	}

	return song, nil
}

func (s *SongService) Search(ctx context.Context, input SearchSongInput) ([]entity.Song, error) {
	songs, err := s.songRepo.Search(ctx, input.Filters, input.OrderBy, input.Offset, input.Limit)
	if err != nil {
		log.Errorf("SongService.Search - s.songRepo.Search: %v", err)
		return []entity.Song{}, ErrCannotGetSong
	}

	return songs, nil
}

func (s *SongService) GetText(ctx context.Context, input GetTextInput) ([]string, int, error) {
	count, err := s.coupletRepo.GetCoupletsCount(ctx, input.SongId)
	if err != nil {
		log.Errorf("SongService.GetText - s.coupletRepo.GetCoupletsCount: %v", err)
		return []string{}, 0, ErrCannotGetText
	}

	couplets, err := s.coupletRepo.GetBySongId(ctx, input.SongId, input.Offset, input.Limit)
	if err != nil {
		log.Errorf("SongService.GetText - s.coupletRepo.GetBySongId: %v", err)
		return []string{}, 0, ErrCannotGetText
	}

	var text []string
	for _, couplet := range couplets {
		text = append(text, couplet.Text)
	}

	return text, count, nil
}

func (s *SongService) Update(ctx context.Context, songId int, input UpdateSongInput) error {
	return nil
}

func (s *SongService) Delete(ctx context.Context, songId int) error {
	err := s.songRepo.DeleteById(ctx, songId)
	if err != nil {
		log.Errorf("SongService.Delete - s.songRepo.DeleteById: %v", err)
		return ErrCannotDeleteSong
	}

	return nil
}
