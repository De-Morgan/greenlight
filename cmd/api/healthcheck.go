package main

import (
	"github.com/gin-gonic/gin"
)

// healthcheckHandler writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) healthcheckHandler(c *gin.Context) {
	js := envelope{
		"environment": app.config.env,
		"version":     version,
	}
	app.successResponse(c, js)
}
