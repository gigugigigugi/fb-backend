package router

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the routes for the application on the given Gin engine.
func SetupRouter(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "WELCOME",
			"env":        "idx",
			"go_version": runtime.Version(),
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"env":        "idx",
			"go_version": runtime.Version(),
		})
	})
}
