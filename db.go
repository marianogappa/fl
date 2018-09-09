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
	ItemName string   `json:"itemName"` // Chose to kept the sqlite schema names untouched
	Location location `json:"location"` // Usually, in non-greenfield one doesn't get to change the schemas
	ItemURL  string   `json:"itemURL"`  // Having said that, an `item` should have a `name`, not an `itemName`
	ImgURLs  []string `json:"imgURLs"`
	// Score    float64  `json:"_score"`
}

const mapping = `
{
	"mappings":{
		"item":{
			"properties":{
				"name":{
					"type":"text",
					"analyzer": "english"
				},
				"location":{
					"type":"geo_point"
				},
				"url":{
					"type":"text",
					"analyzer": "english"
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
	for i := 1; i <= 10; i++ { // Try up to 10 times, because Elasticsearch takes a while to become online
		client, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(url), elastic.SetBasicAuth(user, pass))
		if err == nil {
			break
		}
		log.Printf("newDB: retrying (%v/10) in 1 sec because: %v\n", i, err)
		time.Sleep(1 * time.Second)
	}
	for err != nil {
		return db{}, fmt.Errorf("newDB: could not connect to ES cluster after 10 retries because: %v", err)
	}
	return db{client, index}, nil
}

// mustReplaceIndex deletes db.index if exists, recreates the index and bulk inserts all items
func (db db) mustReplaceIndex(items []item) {
	if err := db.replaceIndex(items); err != nil {
		log.Fatal(err)
	}
}

// replaceIndex deletes db.index if exists, recreates the index and bulk inserts all items
func (db db) replaceIndex(items []item) error {
	exists, err := db.client.IndexExists(db.index).Do(context.Background())
	if err != nil {
		return fmt.Errorf("replaceIndex: couldn't check if index exists: %v", err)
	}
	if exists {
		res, err := db.client.DeleteIndex(db.index).Do(context.Background())
		if res == nil || !res.Acknowledged {
			err = fmt.Errorf("DeleteIndex(%v) wasn't acknowledged by ES", db.index)
		}
		if err != nil {
			return fmt.Errorf("replaceIndex: couldn't delete index: %v", err)
		}
	}
	res, err := db.client.CreateIndex(db.index).BodyString(mapping).Do(context.Background())
	if res == nil || !res.Acknowledged {
		err = fmt.Errorf("CreateIndex(%v) wasn't acknowledged by ES", db.index)
	}
	if err != nil {
		return fmt.Errorf("replaceIndex: couldn't create index: %v", err)
	}
	if err := db.bulkInsertItems(items); err != nil {
		return err
	}
	return nil
}

// Note that bulkInsertItems is only meant to be called once. Otherwise, doc ids will collide.
// This can be mitigated with a different id strategy, but this method is just a convenience feature for reviewing.
func (db db) bulkInsertItems(items []item) error {
	bulkRequest := db.client.Bulk()
	for i, item := range items {
		req := elastic.NewBulkIndexRequest().Index(db.index).Type("item").Id(strconv.Itoa(i)).Doc(item)
		bulkRequest = bulkRequest.Add(req)
	}
	bulkResponse, err := bulkRequest.Do(context.Background())
	if err != nil {
		return fmt.Errorf("bulkInsertItems: couldn't do bulk insert: %v", err)
	}
	if bulkResponse != nil && bulkResponse.Errors {
		return fmt.Errorf("bulkInsertItems: bulk insert had errors")
	}
	if _, err := db.client.Refresh(db.index).Do(context.Background()); err != nil { // force instantly searchable
		return fmt.Errorf("bulkInsertItems: index refresh had error: %v", err)
	}
	return nil
}

func (db db) search(searchTerm string, loc location) ([]item, error) {
	var (
		items = make([]item, 0)
		q     = elastic.NewFunctionScoreQuery()
	)
	q.Query(elastic.NewMultiMatchQuery(searchTerm, "itemName", "itemURL").Analyzer("english"))
	q.AddScoreFunc(elastic.NewGaussDecayFunction().FieldName("location").Origin(loc).Offset("2km").Scale("3km"))

	searchResult, err := db.client.Search().Index(db.index).Query(q).From(0).Size(20).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("search: error executing search query: %v", err)
		log.Println(err)
		return items, err
	}

	for _, hit := range searchResult.Hits.Hits {
		var it item
		if err := json.Unmarshal(*hit.Source, &it); err != nil {
			err = fmt.Errorf("search: error unmarshalling search query result: %v", err)
			log.Println(err)
			return items, err
		}
		// it.Score = *hit.Score
		items = append(items, it)
	}

	return items, nil
}
