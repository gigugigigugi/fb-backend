package router

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"env":        "idx",
			"go_version": runtime.Version(),
		})
	})

	return r
}
