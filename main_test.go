package main

import (
	"fmt"
	"net/http"
	"os"
	"log"

	"github.com/joho/godotenv"
	handler "emerald-moss-api/api"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("VANA Backend running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(handler.Handler)))
}
