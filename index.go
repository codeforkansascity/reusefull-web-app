package main

import (
	"log"
	"net/http"
)

type IndexPageData struct {
	User User
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)
	log.Println(user)
	t.ExecuteTemplate(w, "index.tmpl", IndexPageData{
		User: user,
	})

}
