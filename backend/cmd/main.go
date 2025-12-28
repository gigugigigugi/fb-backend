package main

import (
	"football-backend/common/logger"
	"football-backend/internal/middleware"
	"football-backend/internal/router"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger based on environment
	env := os.Getenv("GIN_MODE")
	if env == "" {
		env = "dev"
	}
	logger.Init(env)

	// Set Gin mode
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	r := gin.New()

	// Register middleware
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery()) // Recovery middleware should be after the logger

	// Setup routes
	router.SetupRouter(r) // Pass the engine to the router setup function

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Server starting", slog.String("port", port))

	// Start server
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
