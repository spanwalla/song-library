package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	_ "github.com/spanwalla/song-library/internal/entity" // for swagger docs
	"github.com/spanwalla/song-library/internal/service"
	"github.com/spanwalla/song-library/pkg/query"
)

type songRoutes struct {
	songService service.Song
}

type songIdInput struct {
	Id int `param:"id" validate:"number,gt=0"`
}

type insertSongInput struct {
	Group string `json:"group" validate:"required,max=128" example:"The Cure"`
	Song  string `json:"song" validate:"required,max=128" example:"Love Song"`
}

type updateSongInput struct {
	Id          int    `param:"id" validate:"number,gt=0"`
	Group       string `json:"group" validate:"required,max=128" example:"Hannah"`
	Song        string `json:"song" validate:"required,max=128" example:"Best Compilation"`
	Link        string `json:"link" validate:"required,max=128,uri" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
	ReleaseDate string `json:"releaseDate" validate:"required,date" example:"2006-06-22"`
}

type updateSongTextInput struct {
	Id   int    `param:"id" validate:"number,gt=0"`
	Text string `json:"text" validate:"required" example:"I can do\nit easily\n\nNew couplet.\n\nAnother one."`
}

func newSongRoutes(g *echo.Group, songService service.Song) {
	r := &songRoutes{songService: songService}

	g.GET("", r.searchSongs)
	g.GET("/:id", r.getSong)
	g.GET("/:id/text", r.getSongText)
	g.DELETE("/:id", r.deleteSong)
	g.PUT("/:id", r.putSong)
	g.PUT("/:id/text", r.putSongText)
	g.POST("", r.insertSong)
}

// @Description Search songs with filters
// @Summary Search songs
// @Param filter[<name>] query string false "Filters, can be multiple" example(Muse)
// @Param order_by query string false "List of sort criteria. Direction will set to asc if it is not stated" example(song:asc,group:desc,release_date)
// @Param offset query int false "Offset" default(0) minimum(0) example(10)
// @Param limit query int false "Limit" default(5) minimum(1) maximum(10) example(10)
// @Produce json
// @Success 200 {array} entity.Song
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs [get]
func (r *songRoutes) searchSongs(c echo.Context) error {
	q := query.NewParams(c.QueryParams())
	q.ParseFilters()
	q.ParseSortCriteria()
	q.ParsePagination()

	var orderBy [][]string
	for _, criteria := range q.SortCriteria {
		orderBy = append(orderBy, []string{criteria.Field, criteria.Order})
	}

	songs, err := r.songService.Search(c.Request().Context(), service.SearchSongInput{
		Filters: q.Filters,
		OrderBy: orderBy,
		Offset:  q.Offset,
		Limit:   q.Limit,
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return err
	}

	return c.JSON(http.StatusOK, songs)
}

// @Description Get song by id
// @Summary Get song by id
// @Param id path int true "Song ID" minimum(1) example(2)
// @Produce json
// @Success 200 {object} entity.Song
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id} [get]
func (r *songRoutes) getSong(c echo.Context) error {
	var input songIdInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	song, err := r.songService.Get(c.Request().Context(), input.Id)
	if err != nil {
		if errors.Is(err, service.ErrSongNotFound) {
			newErrorResponse(c, http.StatusNotFound, err.Error())
		} else {
			newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		}
		return err
	}

	return c.JSON(http.StatusOK, song)
}

// @Description Get song text with pagination by couplets
// @Summary Get song text
// @Param id path int true "Song ID" minimum(1) example(2)
// @Param offset query int false "Offset" default(0) minimum(0) example(10)
// @Param limit query int false "Limit" default(5) minimum(1) maximum(10) example(10)
// @Produce json
// @Success 200 {object} v1.songRoutes.getSongText.response
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id}/text [get]
func (r *songRoutes) getSongText(c echo.Context) error {
	var input songIdInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	q := query.NewParams(c.QueryParams())
	q.ParsePagination()

	couplets, count, err := r.songService.GetText(c.Request().Context(), service.GetTextInput{
		SongId: input.Id,
		Offset: q.Offset,
		Limit:  q.Limit,
	})
	if err != nil {
		if errors.Is(err, service.ErrSongNotFound) {
			newErrorResponse(c, http.StatusNotFound, err.Error())
		} else {
			newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		}
		return err
	}

	type response struct {
		Text  []string `json:"text"`
		Count int      `json:"count"`
	}

	return c.JSON(http.StatusOK, response{
		Text:  couplets,
		Count: count,
	})
}

// @Description Delete song by id
// @Summary Delete song
// @Param id path int true "Song ID" minimum(1) example(2)
// @Success 200
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id} [delete]
func (r *songRoutes) deleteSong(c echo.Context) error {
	var input songIdInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err := r.songService.Delete(c.Request().Context(), input.Id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @Description Edit song by id
// @Summary Edit song
// @Param id path int true "Song ID" minimum(1) example(2)
// @Param song body updateSongInput true "JSON-body"
// @Accept json
// @Success 200
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id} [put]
func (r *songRoutes) putSong(c echo.Context) error {
	var input updateSongInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	releaseDate, err := time.Parse("2006-01-02", input.ReleaseDate)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid date format")
		return err
	}

	err = r.songService.Update(c.Request().Context(), input.Id, service.UpdateSongInput{
		Name:        input.Song,
		Group:       input.Group,
		Link:        input.Link,
		ReleaseDate: releaseDate,
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @Description Edit song text by id
// @Summary Edit song text
// @Param id path int true "Song ID" example(2)
// @Param text body updateSongTextInput true "New song text. Each couplet is separated by double newline symbols."
// @Accept json
// @Success 200
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id}/text [put]
func (r *songRoutes) putSongText(c echo.Context) error {
	var input updateSongTextInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err := r.songService.UpdateText(c.Request().Context(), input.Id, input.Text)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @Description Add new song
// @Summary Add new song
// @Param group body insertSongInput true "Short song info"
// @Accept json
// @Success 201
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs [post]
func (r *songRoutes) insertSong(c echo.Context) error {
	var input insertSongInput

	if err := c.Bind(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return err
	}

	if err := c.Validate(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err := r.songService.Insert(c.Request().Context(), service.InsertSongInput{
		Group: input.Group,
		Song:  input.Song,
	})
	if err != nil {
		if errors.Is(err, service.ErrCannotInsertSong) || errors.Is(err, service.ErrCannotInsertCouplets) ||
			errors.Is(err, service.ErrCannotGetSongInfo) {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		} else {
			newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		}
		return err
	}

	return c.NoContent(http.StatusCreated)
}
