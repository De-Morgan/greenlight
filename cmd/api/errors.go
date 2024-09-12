package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) logError(_ *gin.Context, err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(c *gin.Context, status int, err error) {
	app.JSON(c, status, envelope{
		"status":  "error",
		"message": err.Error(),
	})
}
func (app *application) serverErrorResponse(c *gin.Context, err error) {
	app.logError(c, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(c, http.StatusInternalServerError, errors.New(message))
}

func (app *application) badRequestResponse(c *gin.Context, err error) {
	app.errorResponse(c, http.StatusBadRequest, err)
}

func (app *application) notFoundResponse(c *gin.Context) {
	message := "the requested resource could not be found"
	app.errorResponse(c, http.StatusNotFound, errors.New(message))
}
func (app *application) editConflictResponse(c *gin.Context) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(c, http.StatusConflict, errors.New(message))
}

func (app *application) failedValidationResponse(c *gin.Context, errors map[string]string) {
	app.JSON(c, http.StatusUnprocessableEntity, envelope{
		"status": "failed",
		"errors": errors,
	})
}

func handleJsonDecodeError(err error) error {
	//There is a syntax problem with the JSON being decoded.
	var syntaxError *json.SyntaxError

	//A JSON value is not appropriate for the destination Go type.
	var unmarshalTypeError *json.UnmarshalTypeError

	switch {
	case errors.Is(err, nil):
		return nil
	case errors.As(err, &syntaxError):
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
		}
		return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")

	case err.Error() == "http: request body too large":
		return fmt.Errorf("body must not be larger than %d bytes", maxRequestBodySize)

	default:
		return err
	}

}
