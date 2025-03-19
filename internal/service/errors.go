package service

import "errors"

var (
	ErrCannotGetSongInfo    = errors.New("cannot get song info from external sources")
	ErrCannotInsertSong     = errors.New("cannot insert song")
	ErrCannotInsertCouplets = errors.New("cannot insert couplets")
	ErrSongNotFound         = errors.New("song not found")
	ErrCannotGetSong        = errors.New("cannot get song")
	ErrCannotGetText        = errors.New("cannot get text")
	ErrCannotDeleteSong     = errors.New("cannot delete song")
)
