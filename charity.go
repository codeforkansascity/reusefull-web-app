package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Charity struct {
	Id               int
	Name             string
	Phone            string
	Website          string
	Email            string
	Religious        bool
	Pickup           bool
	Resell           bool
	AmazonWishlist   string
	GoodItems        bool
	CashDonationLink string
	VolunteerSignup  string
	Address          string
	ZipCode          string
	Lat              string
	Long             string
}

type ErrorPage struct {
	ErrorCode int
	Error     string
}

type ListCharityPageData struct {
	Charities []Charity
}

func ListCharities(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select id, name, religious, pickup, resell, good_items, address, zip_code, phone from charity")
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
		charity := Charity{}
		err := rows.Scan(
			&charity.Id,
			&charity.Name,
			&charity.Religious,
			&charity.Pickup,
			&charity.Resell,
			&charity.GoodItems,
			&charity.Address,
			&charity.ZipCode,
			&charity.Phone,
		)
		if err != nil {
			log.Println(err)
			t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
				ErrorCode: 500,
				Error:     err.Error(),
			})
			return
		}
		charities = append(charities, charity)
	}

	t.ExecuteTemplate(w, "charityList.tmpl", ListCharityPageData{
		Charities: charities,
	})
}

func ViewCharity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 400,
			Error:     err.Error(),
		})
		return
	}

	charity := Charity{}
	err = db.QueryRow("select name, religious, pickup, resell, good_items, address, zip_code, phone from charity where id = ?", id).Scan(
		&charity.Name,
		&charity.Religious,
		&charity.Pickup,
		&charity.Resell,
		&charity.GoodItems,
		&charity.Address,
		&charity.ZipCode,
		&charity.Phone,
	)
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	t.ExecuteTemplate(w, "charityView.tmpl", charity)
}

func EditCharity(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", 501)
}

func CharitySignUp1(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "charitySignUp1.tmpl", nil)
}

func CharitySignUp2(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "charitySignUp2.tmpl", nil)
}

func CharitySignUp3(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "charitySignUp3.tmpl", nil)
}
func CharitySignUpComplete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", 501)
	// t.ExecuteTemplate(w, "charitySignUp.tmpl", nil)
}
