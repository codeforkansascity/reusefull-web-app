package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func AdminView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	if !user.Admin {
		http.Error(w, "Forbidden", 403)
		return
	}

	rows, err := db.Query("select id, name, pickup, dropoff, address, phone, logo_url, user_id from charity where approved is null order by id")
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     err.Error(),
		})
		return
	}
	defer rows.Close()

	charities := []Charity{}
	for rows.Next() {
		logoURL := sql.NullString{}
		pickup := sql.NullBool{}
		dropoff := sql.NullBool{}
		userID := sql.NullString{}

		charity := Charity{}
		err := rows.Scan(
			&charity.Id,
			&charity.Name,
			&pickup,
			&dropoff,
			&charity.Address,
			&charity.Phone,
			&logoURL,
			&userID,
		)
		if err != nil {
			log.Println(err)
			t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
				ErrorCode: 500,
				Error:     err.Error(),
			})
			return
		}
		charity.Pickup = pickup.Bool
		charity.Dropoff = dropoff.Bool
		charity.LogoURL = logoURL.String
		charity.UserID = userID.String
		charities = append(charities, charity)
	}

	ids := []string{}
	for _, c := range charities {
		if len(c.UserID) > 0 {
			ids = append(ids, c.UserID)
		}
	}

	if len(ids) > 0 {
		idList := strings.Join(ids, ",")
		stmt := "select id from user where email_verified = true and id in ('" + idList + "')"
		log.Println(stmt)
		rows, err = db.Query(stmt)
		if err != nil {
			log.Println(err)
			t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
				ErrorCode: 500,
				Error:     err.Error(),
			})
		}
		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			if err != nil {
				log.Println(err)
				t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
					ErrorCode: 500,
					Error:     err.Error(),
				})
			}

			for i, c := range charities {
				if c.UserID == id {
					charities[i].EmailVerified = true
					break
				}
			}
		}
	}

	err = t.ExecuteTemplate(w, "adminView.tmpl", struct {
		User      User
		Charities []Charity
	}{
		User:      user,
		Charities: charities,
	})
	if err != nil {
		log.Println(err)
	}

}

func AdminCharityApprove(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	if !user.Admin {
		http.Error(w, "Forbidden", 403)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "id error", 400)
		return
	}

	_, err = db.Exec("update charity set approved = true where id = ?", id)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

}

func AdminCharityDeny(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	if !user.Admin {
		http.Error(w, "Forbidden", 403)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "id error", 400)
		return
	}

	_, err = db.Exec("update charity set approved = false where id = ?", id)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
}
