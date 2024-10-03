package main

import (
	"hc_shows_backend/server"
	"hc_shows_backend/services"
	"hc_shows_backend/utils"
	"log"
	"net/http"
	"os"
)

func main() {
	firestoreService, close, err := services.NewFirestoreHCShowCalendarService()

	if err != nil {
		log.Fatal(err)
	}
	defer close()

	gomailEmailService := services.NewGomailHcShowCalendarEmailService()

	s, err := server.NewHCShowCalendarServer(firestoreService, gomailEmailService)

	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv(utils.APPLICATION_PORT)
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, s); err != nil {
		log.Fatalf("could not listen on port %d %v", port, err)
	}
}
