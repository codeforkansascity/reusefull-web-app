package main

import (
	"net/http"
)

type IndexPageData struct {
	User User
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)
	t.ExecuteTemplate(w, "index.tmpl", IndexPageData{
		User: user,
	})

}
