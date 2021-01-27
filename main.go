package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	_ "github.com/joho/godotenv/autoload"
)

var (
	ss *sessions.FilesystemStore
	db *sql.DB
	t  = template.Must(template.ParseGlob("templates/*"))

	auth0ClientID     string
	auth0ClientSecret string
)

func main() {
	user, exists := os.LookupEnv("MYSQL_USER")
	if !exists {
		panic("MYSQL_USER not found")
	}

	pass, exists := os.LookupEnv("MYSQL_PASS")
	if !exists {
		panic("MYSQL_PASS not found")
	}

	host, exists := os.LookupEnv("MYSQL_HOST")
	if !exists {
		panic("MYSQL_HOST not found")
	}

	auth0ClientID, exists = os.LookupEnv("AUTH0_CLIENT_ID")
	if !exists {
		panic("AUTH0_CLIENT_ID not found")
	}

	auth0ClientSecret, exists = os.LookupEnv("AUTH0_CLIENT_SECRET")
	if !exists {
		panic("AUTH0_CLIENT_SECRET not found")
	}

	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/reusefull?parseTime=true&timeout=10s", user, pass, host))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	ss = sessions.NewFilesystemStore("", []byte("8dp/Kx2veOxt1RdXMBMvlLbwH6oFJDofQyQ1pPodbjQ"))
	gob.Register(map[string]interface{}{})

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(r, "/assets", filesDir)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/", IndexPage)
		r.Get("/charitylist", ListCharities)
		r.Get("/charity/{id}", ViewCharity)
		r.Get("/charity/{id}/edit", EditCharity)
		r.Get("/charity/signup/step/1", CharitySignUp1)
		r.Get("/charity/signup/step/2", CharitySignUp2)
		r.Get("/charity/signup/step/3", CharitySignUp3)
		r.Post("/charity/signup", CharitySignUpComplete)

		r.Get("/auth/callback", CallbackHandler)
		r.Get("/auth/login", LoginHandler)
		r.Get("/auth/logout", LogoutHandler)
	})

	srv := &http.Server{
		Addr:         ":3000",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      r,
	}
	log.Println("Succesfully started")

	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
