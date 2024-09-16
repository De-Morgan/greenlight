package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"morgan.greenlight.nex/internal/data"
	"morgan.greenlight.nex/internal/validator"
)

const (
	maxRequestBodySize = 1 << 20
)

// Middleware to limit request body size
func setMaxSizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxRequestBodySize))
		c.Next()
	}
}

func enableCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func (app *application) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization // header in the request.
		c.Header("Vary", "Authorization")

		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			app.contextSetUser(c, data.AnonymousUser)
			c.Next()
			return
		}

		headerParts := strings.Fields(authorizationHeader)
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthorizationTokenResponse(c)
			c.Abort()
			return
		}
		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthorizationTokenResponse(c)
			c.Abort()
			return
		}
		user, err := app.models.Users.GetForToken(c, data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthorizationTokenResponse(c)
			default:
				app.serverErrorResponse(c, err)
			}
			c.Abort()
			return
		}
		app.contextSetUser(c, user)
		c.Next()
	}
}

// func (app *application) requireAuthenticatedUser() gin.HandlerFunc {

// 	return func(c *gin.Context) {
// 		user := app.contextGetUser(c)
// 		if user.IsAnonymous() {
// 			app.authenticationRequiredResponse(c)
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }

func (app *application) requireActivatedUser() gin.HandlerFunc {

	return func(c *gin.Context) {
		user := app.contextGetUser(c)
		if !user.Activated {
			app.inactiveAccountResponse(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
func (app *application) requirePermissionMiddleware(permisson data.Permission) gin.HandlerFunc {

	return func(c *gin.Context) {
		app.requireActivatedUser()
		user := app.contextGetUser(c)
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(c, err)
			c.Abort()
			return
		}
		if !permissions.Includes(permisson) {
			app.notPermittedResponse(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// func (app *application) rateLimitMiddleware() gin.HandlerFunc {
// 	type client struct {
// 		limiter  *rate.Limiter
// 		lastSeen time.Time
// 	}
// 	var (
// 		mu      sync.Mutex
// 		clients = make(map[string]*client)
// 	)
// 	go func() {
// 		for {
// 			time.Sleep(time.Minute)
// 			mu.Lock()
// 			for ip, client := range clients {
// 				if time.Since(client.lastSeen) > 3*time.Minute {
// 					delete(clients, ip)
// 				}
// 			}
// 			mu.Unlock()
// 		}
// 	}()
// 	return func(c *gin.Context) {
// 		ip := c.ClientIP()
// 		mu.Lock()
// 		if _, ok := clients[ip]; !ok {
// 			clients[ip] = &client{
// 				limiter: rate.NewLimiter(2, 4),
// 			}
// 		}
// 		clients[ip].lastSeen = time.Now()

// 		if !clients[ip].limiter.Allow() {
// 			mu.Unlock()
// 			app.rateLimitExceededResponse(c)
// 			return
// 		}
// 		mu.Unlock()
// 		c.Next()

// 	}
// }
