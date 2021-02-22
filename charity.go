package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	EmailVerified           bool     `json:"emailVerified"`
}

type ErrorPage struct {
	ErrorCode int
	Error     string
	Image     string
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
	rows, err := db.Query("select id, name, pickup, dropoff, address, phone, logo_url from charity order by name")
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

		charity.Phone = FormatPhone(charity.Phone)
		charity.Pickup = pickup.Bool
		charity.Dropoff = dropoff.Bool
		charity.LogoURL = logoURL.String
		charities = append(charities, charity)
	}

	t.ExecuteTemplate(w, "charityList.tmpl", struct {
		User      User
		Charities []Charity
	}{
		User:      r.Context().Value("user").(User),
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

	// Only admins or the owner of the charity can edit
	user := r.Context().Value("user").(User)
	if user.Admin {
		user.CanEdit = true
	} else if user.LoggedIn {
		user.CanEdit = user.ID == charity.UserID
	}

	t.ExecuteTemplate(w, "charityView.tmpl", struct {
		Charity Charity
		User    User
	}{
		Charity: charity,
		User:    user,
	})
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

	// Only admins or the owner of the charity can edit
	user := r.Context().Value("user").(User)
	if user.LoggedIn {
		user.CanEdit = true
	} else if user.LoggedIn {
		user.CanEdit = user.ID == charity.UserID
	}
	if !user.CanEdit {
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 403,
			Error:     "Forbidden",
			Image:     "https://memegenerator.net/img/instances/59208127/ah-ah-ah-you-didnt-say-the-magic-word.jpg",
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
		User         User
		Charity      Charity
		CharityTypes []CharityType
		ItemTypes    []ItemType
	}{
		User:         user,
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

	user := r.Context().Value("user").(User)
	if !user.Admin {
		http.Error(w, "Forbidden", 403)
		return
	}

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

	err = updateLogo(charity.Logo, id)
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

	_, err = tx.Exec(`update charity set
		name = ?,
		address = ?,
		city = ?,
		state = ?,
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
		charity.City,
		charity.State,
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
	t.ExecuteTemplate(w, "charitySignUp1.tmpl", struct {
		User User
	}{
		User: r.Context().Value("user").(User),
	})
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
		User         User
		CharityTypes []CharityType
	}{
		User:         r.Context().Value("user").(User),
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
		User      User
		ItemTypes []ItemType
	}{
		User:      r.Context().Value("user").(User),
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

	res, err := tx.Exec("insert into charity (name, address, city, state, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, pickup, dropoff, faith, taxid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		charity.Name,
		charity.Address,
		charity.City,
		charity.State,
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

	go func() {
		err := updateLogo(charity.Logo, int(charityID))
		if err != nil {
			log.Println(err)
			return
		}
	}()

	go func() {
		err = sendNewAccountEmail(charity.Email, *ticket.Ticket)
		if err != nil {
			log.Println(err)
			http.Error(w, "email error", 500)
			return
		}
	}()

}

func CharitySignUpThanks(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "charitySignUpThanks.tmpl", struct {
		User User
	}{
		User: r.Context().Value("user").(User),
	})
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

func updateLogo(data string, id int) error {
	if len(data) == 0 {
		return nil
	}

	parts := strings.Split(data, ",")
	header := parts[0]

	var key string
	if strings.Contains(header, "image/png") {
		key = fmt.Sprintf("/charities/%d.png", id)
	} else if strings.Contains(header, "image/jpeg") {
		key = fmt.Sprintf("/charities/%d.jpg", id)
	} else {
		return fmt.Errorf("unsupported image format:" + header)
	}

	buf, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	// Upload the file to S3.
	result, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("reusefull"),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return err
	}

	_, err = db.Exec("update charity set logo_url = ? where id = ?", result.Location, id)
	return err
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
	city := sql.NullString{}
	state := sql.NullString{}

	charity := Charity{}
	err := db.QueryRow("select id, name, address, city, state, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, link_logo, pickup, dropoff, faith, approved, taxid, user_id, logo_url from charity where id = ?", id).Scan(
		&charity.Id,
		&charity.Name,
		&charity.Address,
		&city,
		&state,
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
	charity.City = city.String
	charity.State = state.String
	charity.Phone = FormatPhone(charity.Phone)

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

// Formats charity phone nums for display
func FormatPhone(phone string) string {
	/*
		Some phone nums currently contain non-num chars.
		Once cleaning performed on submission, this regex filter
		probably won't be necessary.
	*/
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Println(err)
	}
	cleanedPhone := reg.ReplaceAllString(phone, "")

	formatted := fmt.Sprintf("(%s) %s-%s", cleanedPhone[0:3], cleanedPhone[3:6], cleanedPhone[6:])
	return formatted
}