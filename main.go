package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"emerald-moss-api/pkg/server"
	"github.com/joho/godotenv"
)

// Local development server. Production runs the same routes through api/index.go on Vercel.
func main() {
	_ = godotenv.Load()

	srv := server.NewFromEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("VANA Backend running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, srv.Handler()))
}
