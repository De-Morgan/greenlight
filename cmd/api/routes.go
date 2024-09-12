package main

import (
	"github.com/gin-gonic/gin"
)

const (
	apiVersion = "/v1"
)

func (app *application) routes() *gin.Engine {

	router := gin.Default()
	router.Use(setMaxSizeMiddleware())
	router.NoRoute(app.notFoundResponse)

	v1 := router.Group(apiVersion)

	v1.GET("/healthcheck", app.healthcheckHandler)
	v1.POST("/movies", app.createMovieHandler)
	v1.GET("/movies", app.listMoviesHandler)
	v1.GET("/movies/:id", app.showMovieHandler)
	v1.PATCH("/movies/:id", app.updateMovieHandler)
	v1.DELETE("/movies/:id", app.deleteMovieHandler)

	return router

}
