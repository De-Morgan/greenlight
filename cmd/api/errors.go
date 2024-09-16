package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) logError(c *gin.Context, err error) {
	app.logger.PrintError(err.Error(), map[string]any{
		"request_method": c.Request.Method,
		"request_url":    c.Request.URL,
	})
}

func (app *application) errorResponse(c *gin.Context, status int, err error) {
	app.json(c, status, envelope{
		"status":  "error",
		"message": err.Error(),
	})
}

func (app *application) invalidCredentialsResponse(c *gin.Context) {
	app.errorResponse(c, http.StatusUnauthorized, errors.New("invalid authentication credentials"))
}
func (app *application) invalidAuthorizationTokenResponse(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	app.errorResponse(c, http.StatusUnauthorized, errors.New("invalid or missing authentication token"))
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

// func (app *application) authenticationRequiredResponse(c *gin.Context) {
// 	message := "you must be authenticated to access this resource"
// 	app.errorResponse(c, http.StatusUnauthorized, errors.New(message))
// }

func (app *application) inactiveAccountResponse(c *gin.Context) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(c, http.StatusForbidden, errors.New(message))
}
func (app *application) notPermittedResponse(c *gin.Context) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(c, http.StatusForbidden, errors.New(message))
}

func (app *application) failedValidationResponse(c *gin.Context, errors map[string]string) {
	app.json(c, http.StatusUnprocessableEntity, envelope{
		"status": "failed",
		"errors": errors,
	})
}

// func (app *application) rateLimitExceededResponse(c *gin.Context) {
// 	message := "rate limit exceeded"
// 	app.errorResponse(c, http.StatusTooManyRequests, errors.New(message))
// }

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
