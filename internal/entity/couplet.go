package entity

type Couplet struct {
	SongId         int    `db:"song_id"`
	SequenceNumber int    `db:"sequence_number"`
	Text           string `db:"couplet_text"`
}
