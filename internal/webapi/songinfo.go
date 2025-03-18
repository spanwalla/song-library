package webapi

import (
	"context"
	"time"
)

type SongInfoBody struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongInfoWebAPI struct {
	// client *http.Client
	URL string
}

func NewSongInfoWebAPI(URL string) *SongInfoWebAPI {
	// return &SongInfoWebAPI{http.DefaultClient, URL}
	return &SongInfoWebAPI{URL}
}

func (siw *SongInfoWebAPI) Get(ctx context.Context, group, song string) (GetSongInfoOutput, error) {
	return GetSongInfoOutput{
		ReleaseDate: time.Time{},
		Text:        "Ooh baby, don't you know I suffer?\\nOoh baby, can\nyou hear me moan?\\nYou caught me under false pretenses\\nHow long\nbefore you let me go?\\n\\nOoh\\nYou set my soul alight\\nOoh\\nYou set\nmy soul alight",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}, nil
}
