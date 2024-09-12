package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/validator"
)

type envelope map[string]any

func (app *application) json(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func (app *application) successResponse(c *gin.Context, data any) {
	ev := envelope{
		"status": "success",
		"data":   data,
	}
	app.json(c, http.StatusOK, ev)
}

func (app *application) decodeJson(body io.Reader, data any) error {
	err := json.NewDecoder(body).Decode(data)
	return handleJsonDecodeError(err)
}

func readStringQuery(c *gin.Context, key string, fallback string) string {
	s := c.Query(key)
	if s == "" {
		return fallback
	}
	return s
}
func readStringArrayQuery(c *gin.Context, key string, fallback []string) []string {
	s := c.Query(key)
	if s == "" {
		return fallback
	}
	ss := strings.Split(s, ",")
	if len(ss) == 0 {
		return fallback
	}
	return ss
}

func readIntQuery(c *gin.Context, key string, fallback int, v *validator.Validator) int {
	s := c.Query(key)
	if s == "" {
		return fallback
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return fallback
	}
	return i
}

func validateFilters(c *gin.Context, v *validator.Validator, f *data.Filters) {
	v.Check(f.Page < 0, "page", "must be greater than zero")
	v.Check(f.Page >= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.Limit < 0, "limit", "must be greater than zero")
	v.Check(f.Limit >= 100, "limit", "must be a maximum of 100")
	f.Sort = readStringQuery(c, "sort", "")
	f.SearchTerm = readStringQuery(c, "search_term", "")
	if len(f.SortSafelist) != 0 && f.Sort != "" {
		v.Check(!f.SortPresent(), "sort", "invalid sort value")
	}
	f.Page = readIntQuery(c, "page", 1, v)
	f.Limit = readIntQuery(c, "limit", 20, v)

}
