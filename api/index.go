package handler

import (
	"net/http"
	"sync"

	"emerald-moss-api/pkg/server"
)

var (
	api  http.Handler
	once sync.Once
)

// Handler is the Vercel serverless entry point. The server is built once per
// warm instance; it reconnects to the database on its own if needed.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() { api = server.NewFromEnv().Handler() })
	api.ServeHTTP(w, r)
}
