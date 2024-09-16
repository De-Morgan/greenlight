package main

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(c *gin.Context) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.decodeJson(c.Request.Body, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()
	user, err := app.models.Users.GetByEmail(ctx, input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(c)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.createdResponse(c, envelope{"authentication_token": token})
}
