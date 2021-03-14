package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	hereKey = "EvuC4m1owocrq4WHrdjLNDTyMxfyysf8b00StffLGjk"
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

func geocode(address string) (*Location, error) {
	req, err := http.NewRequest("GET", "https://geocode.search.hereapi.com/v1/geocode", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", "723 Cottonwood Ln, Liberty MO 64068")
	q.Add("apiKey", hereKey)
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
