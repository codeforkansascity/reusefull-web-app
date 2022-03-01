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

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/hyprcubd/dgraphql"
	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/auth0.v5/management"
)

var (
	ss            *sessions.CookieStore
	sesSvc        *ses.SES
	s3Uploader    *s3manager.Uploader
	db            *sql.DB
	authenticator *Authenticator
	t             = template.Must(template.ParseGlob("templates/*"))

	auth0ClientID     string
	auth0ClientSecret string
	auth0RedirectURL  string
	auth0LogoutURL    string
	hereToken         string
	dgraphToken       string
	authManager       *management.Management

	dc *dgraphql.Client
)

const (
	auth0DBConnectionID = "con_p2bySCWPL9MKhmH1"
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

	auth0RedirectURL, exists = os.LookupEnv("AUTH0_REDIRECT_URL")
	if !exists {
		panic("AUTH0_REDIRECT_URL not found")
	}

	auth0LogoutURL, exists = os.LookupEnv("AUTH0_LOGOUT_URL")
	if !exists {
		panic("AUTH0_LOGOUT_URL not found")
	}

	hereToken, exists = os.LookupEnv("HERE_TOKEN")
	if !exists {
		panic("HERE_TOKEN not found")
	}

	dgraphToken, exists = os.LookupEnv("DGRAPH_TOKEN")
	if !exists {
		panic("DGRAPH_TOKEN not found")
	}

	dc = dgraphql.New("https://reusefull.us-west-2.aws.cloud.dgraph.io/graphql", dgraphToken)

	var err error
	authManager, err = management.New("reusefull.us.auth0.com", management.WithClientCredentials(auth0ClientID, auth0ClientSecret))
	if err != nil {
		panic(err)
	}

	authenticator, err = NewAuthenticator()
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	// Simple Email Service
	sesSvc = ses.New(sess)

	// S3 Uploader
	s3Uploader = s3manager.NewUploader(sess)

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/reusefull?parseTime=true&timeout=10s", user, pass, host))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	ss = sessions.NewCookieStore([]byte("8dp/Kx2veOxt1RdXMBMvlLbwH6oFJDofQyQ1pPodbjQ"))
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
		r.Put("/api/v1/charity/{id}", UpdateCharity)

		r.Get("/auth/callback", CallbackHandler)
		r.Get("/auth/login", LoginHandler)
		r.Get("/auth/logout", LogoutHandler)

		r.Get("/charity/signup/step/1", CharitySignUp1)
		r.Get("/charity/signup/step/2", CharitySignUp2)
		r.Get("/charity/signup/step/3", CharitySignUp3)
		r.Get("/charity/signup/thankyou", CharitySignUpThanks)

		r.Get("/donate", Donate)
		r.Get("/donate/results", DonateSearchResults)

		r.Get("/admin", AdminView)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/charity/register", CharityRegister)
		r.Get("/charity/types", GetCharityTypes)
		r.Get("/charity/{id}", GetCharity)

		r.Post("/auth0/cb", ChangePasswordCallback)

		r.Post("/donate/search", DonateSearch)

		r.Post("/charity/{id}/contact", CharityContact)

		// Admin only api
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware)
			r.Put("/charity/{id}/approve", AdminCharityApprove)
			r.Put("/charity/{id}/deny", AdminCharityDeny)
		})
	})

	srv := &http.Server{
		Addr:         ":80",
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
