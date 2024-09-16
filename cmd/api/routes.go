package main

import (
	"expvar"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
)

const (
	apiVersion = "/v1"
)

func (app *application) routes() *gin.Engine {

	router := gin.Default()
	router.Use(enableCORSMiddleware())
	router.Use(setMaxSizeMiddleware())
	router.Use(app.authenticate())
	router.NoRoute(app.notFoundResponse)

	v1 := router.Group(apiVersion)
	movieReadRoute := router.Group(apiVersion).Use(app.requirePermissionMiddleware(data.PermissionMovieRead))
	movieWriteRoute := router.Group(apiVersion).Use(app.requirePermissionMiddleware(data.PermissionMovieWrite))

	v1.GET("/healthcheck", app.healthcheckHandler)
	movieReadRoute.GET("/movies", app.listMoviesHandler)
	movieReadRoute.GET("/movies/:id", app.showMovieHandler)
	movieWriteRoute.POST("/movies", app.createMovieHandler)
	movieWriteRoute.PATCH("/movies/:id", app.updateMovieHandler)
	movieWriteRoute.DELETE("/movies/:id", app.deleteMovieHandler)

	//Users route
	v1.POST("/users", app.registerUserHandler)
	v1.PUT("/users/activated", app.activateUserHandler)

	//Tokens route
	v1.POST("/tokens/authentication", app.createAuthenticationTokenHandler)

	//Application metrics
	router.GET("/metrics", func(ctx *gin.Context) {
		expvar.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
	return router

}
