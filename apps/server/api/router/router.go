package router

import (
	"net/http"

	"telem-api-server/api/resource/session"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	// home API
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// add more routes as we continue
	mux.HandleFunc("/sessions", session.SessionsHandler)
	mux.HandleFunc("/sessions/", session.SessionHandler)
	mux.HandleFunc("/sessions/keys", session.SessionKeyHandler)
	return mux
}
