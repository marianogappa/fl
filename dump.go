package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func mustReadCSVFromFile(path string) []item {
	fh, err := os.Open("dump.csv")
	if err != nil {
		log.Fatal(err)
	}
	items, err := readCSV(fh)
	if err != nil {
		log.Fatal(err)
	}
	return items
}

func readCSV(rd io.Reader) ([]item, error) {
	r := csv.NewReader(rd)
	items := make([]item, 0)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return items, err
		}
		if len(row) != 5 {
			return items, fmt.Errorf("Row didn't have 5 columns: %v", row)
		}
		itemName := row[0]
		lat, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return items, fmt.Errorf("%v (%v)", err, row[1])
		}
		lng, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return items, fmt.Errorf("%v (%v)", err, row[2])
		}
		itemURL := row[3]
		imgURLs := make([]string, 0)
		if err := json.Unmarshal([]byte(row[4]), &imgURLs); err != nil {
			return items, fmt.Errorf("%v (%v)", err, row[4])
		}

		items = append(items,
			item{ItemName: itemName, Location: location{Lat: lat, Lon: lng}, ItemURL: itemURL, ImgURLs: imgURLs},
		)
	}
	return items, nil
}
