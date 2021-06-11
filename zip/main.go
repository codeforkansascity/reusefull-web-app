package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hyprcubd/dgraphql"
)

type Zip struct {
	ID         string `json:"id"`
	Place      string
	State      string
	StateCode  string
	CountyName string
	Location   Point
}

type Point struct {
	Latitude  string
	Longitude string
}

func main() {
	dc := dgraphql.New("https://reusefull.us-west-2.aws.cloud.dgraph.io/graphql", "NmYwMjE5YjRjMzU4YmZjOTRhNzk0ZTM4NjEyNjQ1ZTQ=")

	f, err := os.Open("zip_geo.csv")
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	i := 0
	for {
		i++
		log.Println("Processing record ", i)
		record, err := r.Read()
		if err != nil {
			panic(err)
		}
		if strings.HasPrefix(record[0], "//") {
			continue
		}

		// US,99553,Akutan,Alaska,AK,Aleutians East,13,,,54.143,-165.7854,1
		z := Zip{
			ID:        record[1],
			Place:     record[2],
			State:     record[3],
			StateCode: record[4],
			Location: Point{
				Latitude:  record[9],
				Longitude: record[10],
			},
		}
		log.Println(record)
		log.Println(z)

		err = dc.RawQuery(context.Background(), fmt.Sprintf(`
			mutation MyMutation {
			  addZip(input: {
				  id: "%s",
				  location: {
					  latitude: %s,
					  longitude: %s
					  },
				  countyName: "%s",
				  place: "%s",
				  state: "%s",
				  stateCode: "%s"
				  }) {
			    numUids
			  }
			}`, z.ID, z.Location.Latitude, z.Location.Longitude,
			z.CountyName, z.Place, z.State, z.StateCode), nil)
		if err != nil {
			log.Println(err)
		}

	}

	//
	//

}
