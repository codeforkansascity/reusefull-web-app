package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type DonateSearchRequest struct {
	Zip           string   `json:"zip"`
	OrgSize       string   `json:"orgSize"`
	ItemTypes     []string `json:"itemTypes"`
	CharityTypes  []string `json:"charityTypes"`
	PickupDropoff string   `json:"pickupDropoff"`
	Proximity     string   `json:"proximity"`
}

func DonateStep1(w http.ResponseWriter, r *http.Request) {
	types, err := getItemTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}
	log.Println(types)

	t.ExecuteTemplate(w, "donateSearch1.tmpl", struct {
		ItemTypes []ItemType
	}{
		ItemTypes: types,
	})
}

func DonateStep2(w http.ResponseWriter, r *http.Request) {
	types, err := getCharityTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}
	log.Println(types)

	t.ExecuteTemplate(w, "donateSearch2.tmpl", struct {
		CharityTypes []CharityType
	}{
		CharityTypes: types,
	})
}

func DonateSearchResults(w http.ResponseWriter, r *http.Request) {
	err := t.ExecuteTemplate(w, "donateSearch.tmpl", nil)
	if err != nil {
		log.Println(err)
	}
}

func DonateSearch(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "error reading body", 400)
		return
	}

	req := DonateSearchRequest{}
	err = json.Unmarshal(buf, &req)
	if err != nil {
		log.Println(err)
		http.Error(w, "error parsing json", 400)
		return
	}

	rows, err := db.Query("select c.id, c.name, c.logo_url from charity c")
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	charities := []Charity{}
	for rows.Next() {
		logoURL := sql.NullString{}
		charity := Charity{}
		err = rows.Scan(
			&charity.Id,
			&charity.Name,
			&logoURL,
		)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		charity.LogoURL = logoURL.String
		charities = append(charities, charity)
	}

	data, err := json.Marshal(charities)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	w.Write(data)
}
