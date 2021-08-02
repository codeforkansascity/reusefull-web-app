package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

type Location struct {
	Title      string `json:"title"`
	ID         string `json:"id"`
	ResultType string `json:"resultType"`
	Position   struct {
		Lat float32
		Lng float32
	}
}

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

func Geocode(address string) (*Location, error) {
	if len(hereToken) == 0 {
		return nil, fmt.Errorf("here token not found")
	}

	req, err := http.NewRequest("GET", "https://geocode.search.hereapi.com/v1/geocode", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", address)
	q.Add("apiKey", hereToken)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	log.Println(resp.StatusCode)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 return code %d", resp.StatusCode)
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	output := struct {
		Items []Location
	}{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return nil, err
	}

	return &output.Items[0], nil
}

// Distance uses the Harversine (great circle distance) formula
// for determining the distance between two coordinates
// Converted to Go from this answer:
//   https://stackoverflow.com/questions/27928/calculate-distance-between-two-latitude-longitude-points-haversine-formula
func Distance(lat1, lng1, lat2, lng2 float64) float64 {
	earthRadiusKM := float64(6371)

	dLat := deg2rad(lat2 - lat1)
	dLon := deg2rad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKM * c
}

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}
