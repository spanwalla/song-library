package v1

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/spanwalla/song-library/internal/entity"
	"github.com/spanwalla/song-library/internal/service"
	"github.com/spanwalla/song-library/pkg/query"
	"net/http"
)

type songRoutes struct {
	songService service.Song
}

type songIdInput struct {
	Id int `param:"id" validate:"required,number,gt=0"`
}

type insertSongInput struct {
	Group string `json:"group" validate:"required,max=128"`
	Song  string `json:"song" validate:"required,max=128"`
}

func newSongRoutes(g *echo.Group, songService service.Song) {
	r := &songRoutes{songService: songService}

	g.GET("", r.searchSongs)
	g.GET("/:id", r.getSong)
	g.GET("/:id/text", r.getSongText)
	g.DELETE("/:id", r.deleteSong)
	g.PATCH("/:id", r.patchSong)
	g.POST("", r.insertSong)
}

// @Description Search songs with filters
// @Summary Search songs
// @Param filter[<name>] query string false "Example: &filter[name]=Song&filter[group]=Muse"
// @Param order_by query string false "Example: &order_by=name:asc,link:desc,group. Default: asc"
// @Param offset query int false "Offset, default 0"
// @Param limit query int false "Limit, default 5"
// @Produce json
// @Success 200 {array} entity.Song
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs [get]
func (r *songRoutes) searchSongs(c echo.Context) error {
	return c.JSON(http.StatusOK, []entity.Song{})
}

// @Description Get song by id
// @Summary Get song by id
// @Param id path int true "Song ID"
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
// @Param id path int true "Song ID"
// @Param offset query int false "Offset, default 0"
// @Param limit query int false "Limit, default 5"
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
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
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
// @Param id path int true "Song ID"
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
// @Param id path int true "Song ID"
// @Param song body string false "Song name"
// @Param group body string false "Group name"
// @Accept json
// @Success 200
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /songs/{id} [patch]
func (r *songRoutes) patchSong(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// @Description Add new song
// @Summary Add new song
// @Param song body string true "Song name"
// @Param group body string true "Group name"
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

func (r *songRoutes) toDo(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}
