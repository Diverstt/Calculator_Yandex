package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Diverstt/Calculator_Yandex/internal/orchestrator"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiServer := orchestrator.NewServer()

	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, apiServer.Router))
}
