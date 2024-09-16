package main

import (
	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
)

const (
	userContextKey = "key-user-context"
)

// The contextSetUser() method returns a new copy of the request with the provided
// User struct added to the context. Note that we use our userContextKey constant as the
// key.
func (app *application) contextSetUser(c *gin.Context, user *data.User) {
	c.Set(userContextKey, user)
}

func (app *application) contextGetUser(c *gin.Context) *data.User {
	user, ok := c.Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
