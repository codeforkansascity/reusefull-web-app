package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		return nil, fmt.Errorf("Geocode non 200 return code %d", resp.StatusCode)
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
