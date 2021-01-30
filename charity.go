package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/sethvargo/go-password/password"
	auth0 "gopkg.in/auth0.v5"
	"gopkg.in/auth0.v5/management"
)

type Charity struct {
	Id               int    `json:"id"`
	Name             string `json:"orgName"`
	ContactName      string `json:"contactName"`
	Phone            string `json:"phone"`
	Website          string `json:"website"`
	Email            string `json:"email"`
	Faith            bool   `json:"faith"`
	Pickup           bool   `json:"pickup"`
	Dropoff          bool   `json:"dropoff"`
	Resell           bool
	AmazonWishlist   string `json:"amazon"`
	GoodItems        bool
	CashDonationLink string `json:"cashDonate"`
	VolunteerSignup  string `json:"volunteer"`
	Address          string `json:"address"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZipCode          string `json:"zip"`
	Lat              string
	Long             string
	LogoURL          string   `json:"logoURL"`
	Logo             string   `json:"logo"`
	Mission          string   `json:"mission"`
	Description      string   `json:"description"`
	ItemTypes        []string `json:"itemTypes"`
	CharityTypes     []string `json:"charityTypes"`
	CharityTypeOther string   `json:"other"`
	Budget           string   `json:"budget"`
	TaxID            string   `json:"taxID"`
}

type ErrorPage struct {
	ErrorCode int
	Error     string
}

type CharityType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ItemType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func ListCharities(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select id, name, faith, pickup, resell, good_items, address, zip_code, phone from charity")
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
			&charity.Faith,
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

	t.ExecuteTemplate(w, "charityList.tmpl", struct {
		Charities []Charity
	}{
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
	err = db.QueryRow("select name, faith, pickup, resell, good_items, address, zip_code, phone from charity where id = ?", id).Scan(
		&charity.Name,
		&charity.Faith,
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
	types, err := getCharityTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	t.ExecuteTemplate(w, "charitySignUp2.tmpl", struct {
		CharityTypes []CharityType
	}{
		CharityTypes: types,
	})
}

func CharitySignUp3(w http.ResponseWriter, r *http.Request) {
	types, err := getItemTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	t.ExecuteTemplate(w, "charitySignUp3.tmpl", struct {
		ItemTypes []ItemType
	}{
		ItemTypes: types,
	})
}

func CharityRegister(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	charity := Charity{}
	err = json.Unmarshal(buf, &charity)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("insert into charity (name, address, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, pickup, faith, taxid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		charity.Name,
		charity.Address,
		charity.ZipCode,
		charity.Phone,
		charity.Email,
		charity.ContactName,
		charity.Mission,
		charity.Description,
		charity.CashDonationLink,
		charity.VolunteerSignup,
		charity.Website,
		charity.AmazonWishlist,
		charity.Pickup,
		charity.Faith,
		charity.TaxID,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	charityID, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	log.Printf("Created charity %d, with %d items and %d types", charityID, len(charity.ItemTypes),
		len(charity.CharityTypes))

	for _, t := range charity.ItemTypes {
		_, err = tx.Exec("insert into charity_item (charity_id, item_id) values (?,?)", charityID, t)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to insert into charity_item", 400)
			return
		}
	}

	for _, t := range charity.CharityTypes {
		_, err = tx.Exec("insert into charity_type (charity_id, type_id) values (?,?)", charityID, t)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to insert into charity_type", 400)
			return
		}
	}

	pw, err := password.Generate(32, 10, 10, false, false)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create random password", 400)
		return
	}

	user := management.User{
		Email:      &charity.Email,
		Connection: auth0.String("Database"),
		Password:   auth0.String(pw),
		// Don't verify email here, they will on the password reset
		EmailVerified: auth0.Bool(false),
		VerifyEmail:   auth0.Bool(false),
	}
	err = authManager.User.Create(&user)
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			http.Error(w, "user already exists", 409)
			return
		}
		log.Println(err)
		http.Error(w, "failed to create user", 503)
		return
	}
	userID := *user.ID

	_, err = tx.Exec("insert into user (id) values (?)", userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to insert user record", 500)
		return
	}

	_, err = tx.Exec("update charity set user_id=? where id=?", userID, charityID)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to update charity with user", 500)
		return
	}

	ticket := management.Ticket{
		Email:               &charity.Email,
		ConnectionID:        auth0.String(auth0DBConnectionID),
		MarkEmailAsVerified: auth0.Bool(true),
	}
	err = authManager.Ticket.ChangePassword(&ticket)
	if err != nil {
		log.Println("change password error", err)
		http.Error(w, "failed to send change password", 500)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		http.Error(w, "database error", 500)
		return
	}

	err = sendNewAccountEmail(charity.Email, *ticket.Ticket)
	if err != nil {
		log.Println(err)
		http.Error(w, "email error", 500)
		return
	}

}

func CharitySignUpThanks(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "charitySignUpThanks.tmpl", nil)
}

func GetCharityTypes(w http.ResponseWriter, r *http.Request) {
	types, err := getCharityTypes()
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	out, err := json.Marshal(types)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	w.Write(out)
}

func getCharityTypes() ([]CharityType, error) {
	rows, err := db.Query("select id, name from types")
	if err != nil {
		return nil, err
	}

	types := []CharityType{}
	for rows.Next() {
		t := CharityType{}
		err = rows.Scan(&t.Id, &t.Name)
		if err != nil {
			return nil, err
		}

		types = append(types, t)
	}
	return types, nil
}

func getItemTypes() ([]ItemType, error) {
	rows, err := db.Query("select id, name from item")
	if err != nil {
		return nil, err
	}

	types := []ItemType{}
	for rows.Next() {
		t := ItemType{}
		err = rows.Scan(&t.Id, &t.Name)
		if err != nil {
			return nil, err
		}

		types = append(types, t)
	}
	return types, nil
}
