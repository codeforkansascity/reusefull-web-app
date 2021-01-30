package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

type PostChangePasswordHook struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

const (
	auth0Domain         = "reusefull.us.auth0.com"
	auth0CallbackSecret = "1plCMKWXIsi1VpzO7gnC2tk6VnvCBX1dnwHDlTDe3Q"
)

func NewAuthenticator() (*Authenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "https://"+auth0Domain+"/")
	if err != nil {
		log.Printf("failed to get provider: %v", err)
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     auth0ClientID,
		ClientSecret: auth0ClientSecret,
		RedirectURL:  "http://localhost:3000/auth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
		Ctx:      ctx,
	}, nil
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := ss.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	authenticator, err := NewAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := authenticator.Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("no token found: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: auth0ClientID,
	}

	idToken, err := authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)

	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Getting now the userInfo
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id_token"] = rawIDToken
	session.Values["access_token"] = token.AccessToken
	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to logged in page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	session, err := ss.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authenticator, err := NewAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authenticator.Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := ss.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logoutUrl, err := url.Parse("https://" + auth0Domain)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logoutUrl.Path += "/v2/logout"
	parameters := url.Values{}

	var scheme string
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(returnTo.String())
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", auth0ClientID)
	logoutUrl.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := ss.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := User{}

		profile, ok := session.Values["profile"].(map[string]interface{})
		if ok {
			log.Println(profile)
			user.LoggedIn = true
			user.Name = profile["nickname"].(string)
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ChangePasswordCallback(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if len(auth) == 0 {
		http.Error(w, "missing authorization header", 401)
		return
	}
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[1] != auth0CallbackSecret {
		http.Error(w, "unauthorized", 401)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	user := PostChangePasswordHook{}
	err = json.Unmarshal(buf, &user)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	log.Println(user)

	_, err = db.Exec("update set email_verified = true where id = ?", user.UserID)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
}
