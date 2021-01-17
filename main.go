package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
)

var (
	ss *sessions.FilesystemStore
)

func main() {
	// ss = sessions.NewFilesystemStore("", []byte("8dp/Kx2veOxt1RdXMBMvlLbwH6oFJDofQyQ1pPodbjQ"))
	// gob.Register(map[string]interface{}{})

	t := template.Must(template.ParseGlob("templates/*"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// r.Group(func(r chi.Router) {
	// r.Use(AuthMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "index.tmpl", nil)
	})
	// })

	// r.Get("/auth/callback", CallbackHandler)
	// r.Get("/auth/login", LoginHandler)

	srv := &http.Server{
		Addr:         ":3000",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      r,
	}

	log.Println("Succesfully started")

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
