package entity

import "time"

type Song struct {
	Id          int       `db:"id" json:"id"`
	Name        string    `db:"song_name" json:"song"`
	Group       string    `db:"group_name" json:"group"`
	Link        string    `db:"link" json:"link"`
	ReleaseDate time.Time `db:"release_date" json:"releaseDate"`
}
