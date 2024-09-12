package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/validator"
)

func (app *application) createMovieHandler(c *gin.Context) {

	var input MovieRequest

	if err := app.DecodeJson(c.Request.Body, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}

	movie := &data.Movie{
		Title:   *input.Title,
		Year:    *input.Year,
		Runtime: *input.Runtime,
		Genres:  input.Genres,
	}
	if ok, vErr := movie.Validate(); !ok {
		app.failedValidationResponse(c, vErr)
		return
	}
	err := app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("%s/movies/%d", apiVersion, movie.ID))
	app.successResponse(c, envelope{"movie": movie})
}

func (app *application) showMovieHandler(c *gin.Context) {
	pId := c.Param("id")
	id, err := strconv.ParseInt(pId, 10, 64)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}

	movie, shouldReturn := app.getMovie(id, c)
	if shouldReturn {
		return
	}

	app.successResponse(c, envelope{"movie": movie})
}

func (app *application) deleteMovieHandler(c *gin.Context) {
	pId := c.Param("id")
	id, err := strconv.ParseInt(pId, 10, 64)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}
	shouldReturn := app.deleteMovie(id, c)
	if shouldReturn {
		return
	}
	app.JSON(c, http.StatusNoContent, nil)
}
func (app *application) updateMovieHandler(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}

	var input MovieRequest
	if err := app.DecodeJson(c.Request.Body, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}

	movie, shouldReturn := app.getMovie(id, c)
	if shouldReturn {
		return
	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime

	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if len(input.Genres) != 0 {
		movie.Genres = input.Genres
	}
	if ok, vErr := movie.Validate(); !ok {
		app.failedValidationResponse(c, vErr)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}

	app.successResponse(c, envelope{"movie": movie})
}
func (app *application) listMoviesHandler(c *gin.Context) {
	v := validator.New()
	input := MovieListRequest{}

	// //Filters
	input.Year = int32(readIntQuery(c, "year", 0, v))
	input.Genres = readStringArrayQuery(c, "genres", []string{})
	// //supported Sort list
	input.SortSafelist = []string{"id", "title", "year", "runtime"}
	//Validation
	ValidateFilters(c, v, &input.Filters)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(input.Year, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	app.successResponse(c, envelope{"movies": movies, "metadata": metadata})
}

func (app *application) getMovie(id int64, c *gin.Context) (*data.Movie, bool) {
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return nil, true
	}
	return movie, false
}
func (app *application) deleteMovie(id int64, c *gin.Context) (shouldReturn bool) {
	err := app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		shouldReturn = true
	}
	return
}

type MovieRequest struct {
	Title   *string       `json:"title"`
	Year    *int32        `json:"year"`
	Runtime *data.Runtime `json:"runtime"`
	Genres  []string      `json:"genres"`
}

type MovieListRequest struct {
	//filter by year
	Year int32 `json:"year"`
	//filter by genres
	Genres []string `json:"genres"`
	data.Filters
}
