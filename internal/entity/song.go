package entity

import "time"

type Song struct {
	Id          int       `db:"id"`
	Name        string    `db:"song_name"`
	Group       string    `db:"group_name"`
	Link        string    `db:"link"`
	ReleaseDate time.Time `db:"release_date"`
}
