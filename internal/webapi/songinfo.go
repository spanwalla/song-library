package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

type SongInfoBody struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongInfoWebAPI struct {
	client  *http.Client
	BaseURL string
}

func NewSongInfoWebAPI(url string) *SongInfoWebAPI {
	return &SongInfoWebAPI{
		BaseURL: url,
		client:  &http.Client{},
	}
}

func (siw *SongInfoWebAPI) Get(ctx context.Context, group, song string) (GetSongInfoOutput, error) {
	const endpoint = "/info"
	baseURL, err := url.Parse(siw.BaseURL + endpoint)
	if err != nil {
		return GetSongInfoOutput{}, err
	}

	params := baseURL.Query()
	params.Set("group", group)
	params.Set("song", song)
	baseURL.RawQuery = params.Encode()

	log.Debugf("SongInfoWebApi.Get - baseURL.String(): %s", baseURL.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - http.NewRequestWithContext: %w", err)
	}

	resp, err := siw.client.Do(req)
	if err != nil {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - siw.client.Do: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Errorf("SongInfoWebAPI.Get - Body.Close(): %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - ReadAll: %w", err)
	}

	var result SongInfoBody
	if err = json.Unmarshal(body, &result); err != nil {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - json.Unmarshal: %w", err)
	}

	parsedDate, err := time.Parse("02.01.2006", result.ReleaseDate)
	if err != nil {
		return GetSongInfoOutput{}, fmt.Errorf("SongInfoWebAPI.Get - time.Parse: %w", err)
	}

	return GetSongInfoOutput{
		ReleaseDate: parsedDate,
		Text:        result.Text,
		Link:        result.Link,
	}, nil
}
