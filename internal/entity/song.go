package entity

import "time"

type Song struct {
	Id          int       `db:"id" json:"id" example:"1"`
	Name        string    `db:"song_name" json:"song" example:"Smells Like Teen Spirit"`
	Group       string    `db:"group_name" json:"group" example:"Nirvana"`
	Link        string    `db:"link" json:"link" example:"https://www.youtube.com/watch?v=JirXTmnItd4"`
	ReleaseDate time.Time `db:"release_date" json:"releaseDate" example:"2002-10-29T00:00:00Z"`
}
