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

	"github.com/RoaringBitmap/roaring"
)

const metersPerMile = 1609

type DonateSearchRequest struct {
	OrgSize        string   `json:"orgSize"`
	ItemTypes      []string `json:"itemTypes"`
	CharityTypes   []string `json:"charityTypes"`
	AnyCharityType bool     `json:"anyCharityType"`
	PickupDropoff  string   `json:"pickupDropoff"`
	Proximity      string   `json:"proximity"`
	Zip            string   `json:"zip"`
	Lat            float64  `json:"lat"`
	Lng            float64  `json:"lng"`
}

type Zip struct {
	ID       string `json:"id"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
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
		CharityTypes []CharityType
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
	err = json.Unmarshal(buf, &req)
	if err != nil {
		log.Println(err)
		http.Error(w, "error parsing json", 400)
		return
	}
	log.Printf("%+v", req)

	// Check if we need to look up the lat/long for a zip
	if req.Lat == 0 && req.Lng == 0 && len(req.Zip) > 0 {
		zipLocation := struct {
			Data struct {
				GetZip *Zip
			}
		}{}
		err = dc.RawQuery(r.Context(), fmt.Sprintf(`
		{
		  getZip(id: "%s") {
		    location {
		      latitude
		      longitude
		    }
		  }
		}`, req.Zip), &zipLocation)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		log.Println("got zip location")
		req.Lat = zipLocation.Data.GetZip.Location.Latitude
		req.Lng = zipLocation.Data.GetZip.Location.Longitude
	}
	locBits := roaring.New()
	if req.Lat != 0 {
		proximity, err := strconv.Atoi(req.Proximity)
		if err != nil {
			log.Println(err)
			http.Error(w, "invalid proximity", 400)
			return
		}
		charityDistance := struct {
			Data struct {
				QueryCharity []struct {
					ID       int
					Location struct {
						Latitude  float64
						Longitude float64
					}
				}
			}
		}{}
		err = dc.RawQuery(r.Context(), fmt.Sprintf(`
			{
			  queryCharity(filter: {Location: {near: {distance: %d, coordinate: {longitude: %f, latitude: %f}}}}) {
			    id: CharityID
				Location {
			      latitude
			      longitude
			    }
			  }
			}`, proximity*metersPerMile, req.Lng, req.Lat), &charityDistance)
		if err != nil {
			log.Println(err)
			http.Error(w, "server error", 500)
			return
		}
		log.Println(charityDistance.Data.QueryCharity)
		if len(charityDistance.Data.QueryCharity) == 0 {
			data, err := json.Marshal([]Charity{})
			if err != nil {
				log.Println(err)
				http.Error(w, "server error", 500)
				return
			}
			w.Write(data)
			return
		}

		for _, charity := range charityDistance.Data.QueryCharity {
			log.Println("charity within location: ", charity.ID, charity.Location)
			locBits.AddInt(charity.ID)
		}
	}

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
	log.Println("itbits:", itBits.String())
	log.Println("locBits:", locBits.String())
	itBits.And(locBits)
	log.Println("itbits:", itBits.String())

	charities := []Charity{}
	if itBits.GetCardinality() > 0 {
		stmt := "select c.id, c.name, c.address, c.city, c.state, c.zip_code, c.phone, c.mission, c.logo_url, c.pickup, c.dropoff, c.lat, c.lng from charity c where c.id in ("

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
		stmt += "and approved is true "

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
			charity := Charity{}
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
				&lat,
				&lng,
			)
			if err != nil {
				log.Println(err)
				http.Error(w, "server error", 500)
				return
			}
			charity.LogoURL = logoURL.String
			charity.Lat = lat.Float64
			charity.Lng = lng.Float64
			if req.Lat != 0 {
				charity.Distance = Distance(req.Lat, req.Lng, charity.Lat, charity.Lng)
			}
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
