package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/RoaringBitmap/roaring"
	"github.com/hyprcubd/reusefull/models"
)

type DonateSearchRequest struct {
	Zip            string   `json:"zip"`
	OrgSize        string   `json:"orgSize"`
	Resell         bool     `json:"resell"`
	Faith          bool     `json:"faith"`
	NewItems       bool     `json:"newItems"`
	ItemTypes      []string `json:"itemTypes"`
	CharityTypes   []string `json:"charityTypes"`
	AnyCharityType bool     `json:"anyCharityType"`
	PickupDropoff  string   `json:"pickupDropoff"`
	Proximity      string   `json:"proximity"`
}

func Donate(w http.ResponseWriter, r *http.Request) {
	it, err := getItemTypes()
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

	err = t.ExecuteTemplate(w, "donate.tmpl", struct {
		User         User
		CharityTypes []models.CharityType
		ItemTypes    []ItemType
	}{
		User:         r.Context().Value("user").(User),
		CharityTypes: ct,
		ItemTypes:    it,
	})
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
	}
}

func DonateSearchResults(w http.ResponseWriter, r *http.Request) {
	err := t.ExecuteTemplate(w, "donateResults.tmpl", struct {
		User User
	}{
		User: r.Context().Value("user").(User),
	})
	if err != nil {
		log.Println(err)
		t.ExecuteTemplate(w, "error.tmpl", ErrorPage{
			ErrorCode: 500,
			Error:     "Server error. We're looking into it!",
		})
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
	log.Println(req)
	err = json.Unmarshal(buf, &req)
	if err != nil {
		log.Println(err)
		http.Error(w, "error parsing json", 400)
		return
	}
	log.Printf("%+v", req)

	itBits := roaring.New()
	if len(req.ItemTypes) > 0 {
		stmt := "select distinct charity_id from charity_item where item_id in (" + strings.Join(req.ItemTypes, ",") + ")"
		rows, err := db.Query(stmt)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				log.Println(err)
				http.Error(w, "server error", 500)
				rows.Close()
				return
			}
			itBits.Add(uint32(id))
		}
		rows.Close()
	}
	log.Println("item types matching:", itBits.String())

	log.Println("any charity ", req.AnyCharityType)
	ctBits := roaring.New()
	if !req.AnyCharityType && len(req.CharityTypes) > 0 {
		stmt := "select distinct charity_id from charity_type where type_id in (" + strings.Join(req.CharityTypes, ",") + ")"
		rows, err := db.Query(stmt)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				log.Println(err)
				http.Error(w, "server error", 500)
				rows.Close()
				return
			}
			ctBits.Add(uint32(id))
		}
		rows.Close()
	}
	log.Println("charity types matching", ctBits.String())

	if !req.AnyCharityType {
		itBits.And(ctBits)
	}
	log.Println(itBits.String())

	charities := []models.Charity{}
	if itBits.GetCardinality() > 0 {
		stmt := "select c.id, c.name, c.address, c.city, c.state, c.zip_code, c.phone, c.mission, c.logo_url, c.pickup, c.dropoff, c.resell, c.new_items, c.lat, c.lng, c.link_website from charity c where c.id in ("

		first := true
		it := itBits.Iterator()
		for it.HasNext() {
			if !first {
				stmt += ","
			}
			first = false
			stmt += fmt.Sprintf("%d", it.Next())
		}
		stmt += ") "

		if req.PickupDropoff == "1" {
			stmt += "and pickup is true "
		} else if req.PickupDropoff == "2" {
			stmt += "and dropoff is true "
		}

		// only filter by resell if not selected (include by default)
		if !req.Resell {
			stmt += "AND (resell IS false OR resell IS NULL) "
		}

		// only filter by faith if checkbox isn't selected (include by default)
		if !req.Faith {
			stmt += "AND (faith IS false OR faith IS NULL) "
		}

		// if selected, only include orgs that require new items
		if req.NewItems {
			stmt += "AND new_items IS true "
		}

		stmt += "and paused is false and approved is true "

		log.Println(stmt)
		rows, err := db.Query(stmt)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		defer rows.Close()

		for rows.Next() {
			logoURL := sql.NullString{}
			lat := sql.NullFloat64{}
			lng := sql.NullFloat64{}
			charity := models.Charity{}
			err = rows.Scan(
				&charity.Id,
				&charity.Name,
				&charity.Address,
				&charity.City,
				&charity.State,
				&charity.ZipCode,
				&charity.Phone,
				&charity.Mission,
				&logoURL,
				&charity.Pickup,
				&charity.Dropoff,
				&charity.Resell,
				&charity.NewItems,
				&lat,
				&lng,
				&charity.Website,
			)
			if err != nil {
				log.Println(err)
				http.Error(w, "server error", 500)
				return
			}
			charity.LogoURL = logoURL.String
			charity.Lat = lat.Float64
			charity.Lng = lng.Float64
			charities = append(charities, charity)
		}
	}

	data, err := json.Marshal(charities)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
		return
	}

	w.Write(data)
}
