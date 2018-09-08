package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/olivere/elastic"
)

type db struct {
	client *elastic.Client
	index  string
}

type location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type item struct {
	ItemName string   `json:"itemName"`
	Location location `json:"location"`
	ItemURL  string   `json:"itemURL"`
	ImgURLs  []string `json:"imgURLs"`
	// Score    float64  `json:"_score"`
}

const mapping = `
{
	"mappings":{
		"item":{
			"properties":{
				"name":{
					"type":"text"
				},
				"location":{
					"type":"geo_point"
				},
				"url":{
					"type":"text"
				},
				"img_urls":{
					"type":"object"
				}
			}
		}
	}
}`

func mustNewDB(url, user, pass, index string) db {
	db, err := newDB(url, user, pass, index)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func newDB(url, user, pass, index string) (db, error) {
	var (
		client *elastic.Client
		err    error
	)
	for i := 1; i <= 10; i++ {
		client, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(url), elastic.SetBasicAuth(user, pass))
		if err == nil {
			break
		}
		log.Printf("Retrying (%v/10) in 1 sec because: %v\n", i, err)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		return db{}, fmt.Errorf("Could not connect to ES cluster after 10 retries because: %v", err)
	}
	return db{client, index}, nil
}

func (db db) mustReplaceIndex(items []item) {
	err := db.replaceIndex(items)
	if err != nil {
		log.Fatal(err)
	}
}

func (db db) replaceIndex(items []item) error {
	exists, err := db.client.IndexExists(db.index).Do(context.Background())
	if err != nil {
		return err
	}
	if exists {
		res1, err := db.client.DeleteIndex(db.index).Do(context.Background())
		if res1 == nil || !res1.Acknowledged {
			err = fmt.Errorf("db: DeleteIndex(%v) wasn't acknowledged by ES", db.index)
		}
		if err != nil {
			return err
		}
	}
	res2, err := db.client.CreateIndex(db.index).BodyString(mapping).Do(context.Background())
	if res2 == nil || !res2.Acknowledged {
		err = fmt.Errorf("db: CreateIndex(%v) wasn't acknowledged by ES", db.index)
	}
	if err != nil {
		return err
	}
	if err := db.put(items); err != nil {
		return err
	}
	return nil
}

func (db db) put(items []item) error {
	bulkRequest := db.client.Bulk()
	for i, item := range items {
		req := elastic.NewBulkIndexRequest().Index(db.index).Type("item").Id(strconv.Itoa(i)).Doc(item)
		bulkRequest = bulkRequest.Add(req)
	}
	bulkResponse, err := bulkRequest.Do(context.Background())
	if err != nil {
		return err
	}
	if bulkResponse != nil && bulkResponse.Errors {
		return fmt.Errorf("db: bulk insert had errors")
	}
	return nil
}

func (db db) search(searchTerm string, loc location) ([]item, error) {
	var items = make([]item, 0)
	q := elastic.NewFunctionScoreQuery()
	q.Query(elastic.NewQueryStringQuery(searchTerm).Field("itemName").Field("itemURL").Field("imgURLs"))
	q.AddScoreFunc(elastic.NewGaussDecayFunction().FieldName("location").Origin(loc).Offset("2km").Scale("3km"))

	searchResult, err := db.client.Search().Index(db.index).Query(q).From(0).Size(20).Do(context.Background())
	if err != nil {
		return items, err
	}

	for _, hit := range searchResult.Hits.Hits {
		var it item
		if err := json.Unmarshal(*hit.Source, &it); err != nil {
			return items, err
		}
		// it.Score = *hit.Score
		items = append(items, it)
	}

	return items, nil
}
