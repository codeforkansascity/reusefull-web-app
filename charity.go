package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	Id                      int      `json:"id"`
	Name                    string   `json:"name"`
	ContactName             string   `json:"contactName"`
	Phone                   string   `json:"phone"`
	Website                 string   `json:"website"`
	Email                   string   `json:"email"`
	Faith                   bool     `json:"faith"`
	Pickup                  bool     `json:"pickup"`
	Dropoff                 bool     `json:"dropoff"`
	Resell                  bool     `json:"resell"`
	AmazonWishlist          string   `json:"amazon"`
	GoodItems               bool     `json:"goodItems"`
	CashDonationLink        string   `json:"cashDonate"`
	VolunteerSignup         string   `json:"volunteer"`
	Address                 string   `json:"address"`
	City                    string   `json:"city"`
	State                   string   `json:"state"`
	ZipCode                 string   `json:"zip"`
	Lat                     string   `json:"lat"`
	Long                    string   `json:"long"`
	LogoURL                 string   `json:"logoURL"`
	Logo                    string   `json:"logo"`
	Mission                 string   `json:"mission"`
	Description             string   `json:"description"`
	ItemTypes               []string `json:"itemTypes"`
	ItemTypeDescriptions    []string `json:"itemTypeDescriptions"`
	CharityTypes            []string `json:"charityTypes"`
	CharityTypeDescriptions []string `json:"charityTypeDescriptions"`
	CharityTypeOther        string   `json:"other"`
	Budget                  string   `json:"budget"`
	TaxID                   string   `json:"taxID"`
	UserID                  string   `json:"userID"`
	Approved                bool     `json:"approved"`
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
	rows, err := db.Query("select id, name, pickup, dropoff, address, phone, logo_url from charity")
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

		charity := Charity{}
		err := rows.Scan(
			&charity.Id,
			&charity.Name,
			&pickup,
			&dropoff,
			&charity.Address,
			&charity.Phone,
			&logoURL,
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

	charity, err := getCharity(id)
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

func GetCharity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "no id param", 400)
		return
	}

	charity, err := getCharity(id)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	out, err := json.Marshal(charity)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	w.Write(out)
}

func EditCharity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 400,
			Error:     err.Error(),
		})
		return
	}

	charity, err := getCharity(id)
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	ct, err := getCharityTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	it, err := getItemTypes()
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
		return
	}

	t.ExecuteTemplate(w, "charityEdit.tmpl", struct {
		Charity      Charity
		CharityTypes []CharityType
		ItemTypes    []ItemType
	}{
		Charity:      charity,
		CharityTypes: ct,
		ItemTypes:    it,
	})
}

func UpdateCharity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "missing id", 400)
		return
	}

	// TODO: check for admin or owner of charity here

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	log.Println(string(buf))

	charity := Charity{}
	err = json.Unmarshal(buf, &charity)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	log.Println(charity)

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`update charity set
		name = ?,
		address = ?,
		zip_code = ?,
		phone = ?,
		email = ?,
		contact_name = ?,
		mission = ?,
		description = ?,
		link_donate_cash = ?,
		link_volunteer = ?,
		link_website = ?,
		link_wishlist = ?,
		pickup = ?,
		dropoff = ?,
		resell = ?,
		faith = ?,
		good_items = ?,
		taxid = ?
		where id = ?
		`,
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
		charity.Dropoff,
		charity.Resell,
		charity.Faith,
		charity.GoodItems,
		charity.TaxID,
		id)
	if err != nil {
		log.Println("db error update", err)
		http.Error(w, "server error", 500)
		return
	}

	// Clear the previous choices
	_, err = tx.Exec("delete from charity_type where charity_id = ? ", id)
	if err != nil {
		log.Println("db error update", err)
		http.Error(w, "server error", 500)
		return
	}

	for _, t := range charity.CharityTypes {
		_, err = tx.Exec("insert into charity_type (charity_id, type_id) values (?,?)", id, t)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to insert into charity_type", 500)
			return
		}
	}

	// Clear the previous choices
	_, err = tx.Exec("delete from charity_item where charity_id = ? ", id)
	if err != nil {
		log.Println("db error update", err)
		http.Error(w, "server error", 500)
		return
	}

	for _, t := range charity.ItemTypes {
		_, err = tx.Exec("insert into charity_item (charity_id, item_id) values (?,?)", id, t)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to insert into charity_item", 500)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("db error update", err)
		http.Error(w, "server error", 500)
		return
	}

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

	res, err := tx.Exec("insert into charity (name, address, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, pickup, dropoff, faith, taxid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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
		charity.Dropoff,
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
		log.Println("user create error", err)
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

func getCharity(id int) (Charity, error) {
	contactName := sql.NullString{}
	mission := sql.NullString{}
	description := sql.NullString{}
	linkLogo := sql.NullString{}
	pickup := sql.NullBool{}
	dropoff := sql.NullBool{}
	faith := sql.NullBool{}
	approved := sql.NullBool{}
	taxID := sql.NullString{}
	userID := sql.NullString{}
	logoURL := sql.NullString{}

	charity := Charity{}
	err := db.QueryRow("select id, name, address, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, link_logo, pickup, dropoff, faith, approved, taxid, user_id, logo_url from charity where id = ?", id).Scan(
		&charity.Id,
		&charity.Name,
		&charity.Address,
		&charity.ZipCode,
		&charity.Phone,
		&charity.Email,
		&contactName,
		&mission,
		&description,
		&charity.CashDonationLink,
		&charity.VolunteerSignup,
		&charity.Website,
		&charity.AmazonWishlist,
		&linkLogo,
		&pickup,
		&dropoff,
		&faith,
		&approved,
		&taxID,
		&userID,
		&logoURL,
	)
	charity.ContactName = contactName.String
	charity.Mission = mission.String
	charity.Description = description.String
	charity.LogoURL = linkLogo.String
	charity.Pickup = pickup.Bool
	charity.Dropoff = dropoff.Bool
	charity.Faith = faith.Bool
	charity.Approved = approved.Bool
	charity.TaxID = taxID.String
	charity.UserID = userID.String
	charity.LogoURL = logoURL.String

	// Get all the associated charity types
	rows, err := db.Query("select ct.type_id, t.name from charity_type ct, types t where ct.type_id = t.id and charity_id =?", id)
	if err != nil {
		return charity, err
	}

	charity.CharityTypes = []string{}
	for rows.Next() {
		var id int
		var desc string
		err = rows.Scan(&id, &desc)
		if err != nil {
			rows.Close()
			return charity, err
		}
		charity.CharityTypes = append(charity.CharityTypes, fmt.Sprintf("%d", id))
		charity.CharityTypeDescriptions = append(charity.CharityTypeDescriptions, desc)
	}
	rows.Close()

	// Get all the associated item types
	rows, err = db.Query("select ci.item_id, i.name from charity_item ci, item i where ci.item_id = i.id and charity_id = ?", id)
	if err != nil {
		return charity, err
	}

	charity.ItemTypes = []string{}
	for rows.Next() {
		var id int
		var desc string
		err = rows.Scan(&id, &desc)
		if err != nil {
			rows.Close()
			return charity, err
		}
		charity.ItemTypes = append(charity.ItemTypes, fmt.Sprintf("%d", id))
		charity.ItemTypeDescriptions = append(charity.ItemTypeDescriptions, desc)
	}
	rows.Close()
	return charity, err
}
