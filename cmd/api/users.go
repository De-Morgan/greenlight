package main

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/validator"
)

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) registerUserHandler(c *gin.Context) {
	var input UserRequest

	if err := app.decodeJson(c.Request.Body, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}
	v := validator.New()

	user := &data.User{
		Name:  input.Name,
		Email: input.Email,
	}

	//Validate input
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	data.ValidateUserName(v, input.Name)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	if err := user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	err := app.models.Users.Insert(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			app.failedValidationResponse(c, map[string]string{
				"email": "a user with this email address already exists",
			})
			return
		default:
			app.serverErrorResponse(c, err)
			return
		}
	}
	err = app.models.Permissions.AddForUser(user.ID, data.PermissionMovieRead)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.logger.PrintInfo("activationToken", map[string]any{
		"token": token.Plaintext,
	})
	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userId":          user.ID,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err.Error(), nil)
			return
		}
	})

	app.successResponse(c, envelope{"user": user})
}

func (app *application) activateUserHandler(c *gin.Context) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.decodeJson(c.Request.Body, &input)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	ctx, cancel := context.WithTimeout(c, 6*time.Second)
	defer cancel()
	user, err := app.models.Users.GetForToken(ctx, data.ScopeActivation, input.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(c, v.Errors)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	user.Activated = true
	err = app.models.Users.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.successResponse(c, envelope{"user": user})
}
