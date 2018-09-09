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
	fh, err := os.Open(path)
	if err != nil {
		log.Fatalf("mustReadCSVFromFile: error opening file: %v", err)
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
			return items, fmt.Errorf("readCSV: error reading record: %v", err)
		}
		if len(row) != 5 {
			return items, fmt.Errorf("readCSV: row didn't have 5 columns: %v", row)
		}
		itemName := row[0]
		lat, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return items, fmt.Errorf("readCSV: error parsing %v as float: %v", row[1], err)
		}
		lng, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return items, fmt.Errorf("readCSV: error parsing %v as float: %v", row[2], err)
		}
		itemURL := row[3]
		imgURLs := make([]string, 0)
		if err := json.Unmarshal([]byte(row[4]), &imgURLs); err != nil {
			return items, fmt.Errorf("readCSV: error parsing %v as []string: %v", row[4], err)
		}

		items = append(items,
			item{ItemName: itemName, Location: location{Lat: lat, Lon: lng}, ItemURL: itemURL, ImgURLs: imgURLs},
		)
	}
	return items, nil
}
