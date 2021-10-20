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
	"golang.org/x/net/context"
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
	Faith                   *bool    `json:"faith"`
	Pickup                  bool     `json:"pickup"`
	Dropoff                 bool     `json:"dropoff"`
	Resell                  *bool    `json:"resell"`
	NewItems                *bool    `json:"newItems"`
	AmazonWishlist          string   `json:"amazon"`
	GoodItems               bool     `json:"goodItems"`
	CashDonationLink        string   `json:"cashDonate"`
	VolunteerSignup         string   `json:"volunteer"`
	Address                 string   `json:"address"`
	City                    string   `json:"city"`
	State                   string   `json:"state"`
	ZipCode                 string   `json:"zip"`
	LogoURL                 string   `json:"logoURL"`
	Logo                    string   `json:"logo"`
	Lat                     float64  `json:"lat"`
	Lng                     float64  `json:"lng"`
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
	Paused                  bool     `json:"paused"`
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

type Message struct {
	Sender string `json:"sender"`
	Body   string `json:"body"`
	Name   string `json:"name"`
}

func ListCharities(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select id, name, pickup, dropoff, address, city, state, zip_code, phone, logo_url
		from charity where approved is true order by name`)
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
			&charity.City,
			&charity.State,
			&charity.ZipCode,
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
		new_items = ?,
		taxid = ?,
		paused = ?
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
		charity.NewItems,
		charity.TaxID,
		charity.Paused,
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

	// Geocode
	go func() {
		address := fmt.Sprintf("%s %s, %s %s", charity.Address, charity.City, charity.State, charity.ZipCode)
		log.Println("Geocoding address ", address)
		loc, err := Geocode(address)
		if err != nil {
			log.Println(err)
			return
		}
		err = dc.RawQuery(context.Background(), fmt.Sprintf(`
			mutation {
				updateCharity(input: {
					filter: {
						CharityID: {eq: %d}
					},
					set: {
						Address: "%s",
						City: "%s",
						State: "%s",
						Zip: "%s",
						FullAddress: "%s",
						Location: {
							longitude: %f,
							latitude: %f
						},
					}
				}) {
					numUids
				}
			}
			`, charity.Id, charity.Address, charity.City, charity.State, charity.ZipCode, loc.Title, loc.Position.Lng, loc.Position.Lat), nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("saved in dgraph")

		_, err = db.Exec("update charity set lat=?, lng=? where id = ?", loc.Position.Lat, loc.Position.Lng, charity.Id)
		if err != nil {
			log.Println(err)
			return
		}

	}()

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

	res, err := tx.Exec("insert into charity (name, address, city, state, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, pickup, dropoff, faith, resell, new_items, taxid, paused) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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
		charity.Resell,
		charity.NewItems,
		charity.TaxID,
		false,
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

	// Geocode
	go func() {
		address := fmt.Sprintf("%s %s, %s %s", charity.Address, charity.City, charity.State, charity.ZipCode)
		log.Println("Geocoding address ", address)
		loc, err := Geocode(address)
		if err != nil {
			log.Println(err)
			return
		}
		err = dc.RawQuery(context.Background(), fmt.Sprintf(`
			mutation MyMutation {
				addCharity(input: {
					CharityID: %d,
					Address: "%s",
					City: "%s",
					State: "%s",
					Zip: "%s",
					Location: {
						longitude: %f,
						latitude: %f
					},
					FullAddress: "%s"}) {
					numUids
				}
			}
			`, charity.Id, charity.Address, charity.City, charity.State, charity.ZipCode,
			loc.Position.Lng, loc.Position.Lat, loc.Title), nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("saved in dgraph")

		_, err = db.Exec("update charity set lat=?, lng=? where id = ?", loc.Position.Lat, loc.Position.Lng, charity.Id)
		if err != nil {
			log.Println(err)
			return
		}

	}()

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
		err = sendAdminNotificationEmail(charity.Name)
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
	resell := sql.NullBool{}
	newItems := sql.NullBool{}
	approved := sql.NullBool{}
	taxID := sql.NullString{}
	userID := sql.NullString{}
	logoURL := sql.NullString{}
	city := sql.NullString{}
	state := sql.NullString{}
	paused := sql.NullBool{}

	charity := Charity{}
	err := db.QueryRow("select id, name, address, city, state, zip_code, phone, email, contact_name, mission, description, link_donate_cash, link_volunteer, link_website, link_wishlist, link_logo, pickup, dropoff, faith, resell, new_items, approved, taxid, user_id, logo_url, paused from charity where id = ?", id).Scan(
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
		&resell,
		&newItems,
		&approved,
		&taxID,
		&userID,
		&logoURL,
		&paused,
	)
	charity.ContactName = contactName.String
	charity.Mission = mission.String
	charity.Description = description.String
	charity.LogoURL = linkLogo.String
	charity.Pickup = pickup.Bool
	charity.Dropoff = dropoff.Bool
	charity.Faith = &faith.Bool
	charity.Resell = &resell.Bool
	charity.NewItems = &newItems.Bool
	charity.Approved = approved.Bool
	charity.TaxID = taxID.String
	charity.UserID = userID.String
	charity.LogoURL = logoURL.String
	charity.City = city.String
	charity.State = state.String
	charity.Phone = FormatPhone(charity.Phone)
	charity.Paused = paused.Bool

	// change relative links to absolute
	charity.Website = convertToAbsoluteURL(charity.Website)
	charity.AmazonWishlist = convertToAbsoluteURL(charity.AmazonWishlist)
	charity.CashDonationLink = convertToAbsoluteURL(charity.CashDonationLink)
	charity.VolunteerSignup = convertToAbsoluteURL(charity.VolunteerSignup)

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

/*
	Sends an email from the contact form on a charity page to the contact email
	associated with that charity's ID.
*/
func CharityContact(w http.ResponseWriter, r *http.Request) {

	// Extract message details from request body
	buf, err := ioutil.ReadAll(r.Body)
	message := Message{}
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}
	err = json.Unmarshal(buf, &message)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	// Get charity's ID from URL params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "id error", 400)
		return
	}

	// Retrieve contact email associated with that charity
	charityEmail := sql.NullString{}
	err = db.QueryRow("select email from charity where id = ?", id).Scan(&charityEmail)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	// Send a formatted message with the sender's name, message, and contact email
	err = sendContactEmail(charityEmail.String, message.Sender, message.Name, message.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

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

func convertToAbsoluteURL(url string) string {
	if len(url) > 4 {
		if strings.Contains(url, "@") {
			url = "mailto:" + url
			return url
		}
		if !strings.Contains(url, "http") {
			url = "http://" + url
		}
	}

	return url
}
