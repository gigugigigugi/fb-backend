package main

import (
	"football-backend/internal/router"
	"log"
	"net/http"
)

func main() {
	r := router.SetupRouter()
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
