package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// Integration test creates a new index for every subtest.
// Tests by making HTTP requests and expecting a response status code and payload.
func TestIntegration(t *testing.T) {
	var (
		db, err = newDB("http://elasticsearch:9200", "elastic", "changeme", "")
		ts      = []struct {
			name               string
			items              string
			httpMethod         string
			endpoint           string
			searchTerm         string
			lat                string
			lon                string
			expected           []item
			expectedStatusCode int
		}{
			{
				name:               "POST method not allowed",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "POST",
				endpoint:           "/search",
				searchTerm:         "camera",
				lat:                "0",
				lon:                "0",
				expected:           []item{},
				expectedStatusCode: http.StatusMethodNotAllowed,
			},
			{
				name:               "/differentEndpoint endpoint not found",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/differentEndpoint",
				searchTerm:         "camera",
				lat:                "0",
				lon:                "0",
				expected:           []item{},
				expectedStatusCode: http.StatusNotFound,
			},
			{
				name:               "empty search term returns Bad Request",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "",
				lat:                "51",
				lon:                "0",
				expected:           []item{},
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				name:               "happy case",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "camera",
				lat:                "51",
				lon:                "0",
				expected:           []item{{"camera", location{51, 0}, "london/camera", []string{}}},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:               "plural match",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "cameras",
				lat:                "51",
				lon:                "0",
				expected:           []item{{"camera", location{51, 0}, "london/camera", []string{}}},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:               "many words, one is similar match",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "video cameras",
				lat:                "51",
				lon:                "0",
				expected:           []item{{"camera", location{51, 0}, "london/camera", []string{}}},
				expectedStatusCode: http.StatusOK,
			},
		}
	)
	if err != nil {
		t.Errorf("can't connect to ES: %v", err)
		t.FailNow()
	}
	defer db.client.Stop()
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			var (
				index      = "test_items_" + randomHash()
				items, err = readCSV(strings.NewReader(tc.items))
			)
			db.index = index
			if err != nil {
				t.Errorf("couldn't read items: %v", err)
				t.FailNow()
			}
			if err := db.replaceIndex(items); err != nil {
				t.Errorf("couldn't replace index: %v", err)
				t.FailNow()
			}
			defer db.deleteIndex()
			var (
				server = httptest.NewServer(http.HandlerFunc(newEndpointHandler(db).ServeHTTP))
				client = http.Client{}
				url    = fmt.Sprintf("%v%v?searchTerm=%v&lat=%v&lng=%v",
					server.URL, tc.endpoint, url.PathEscape(tc.searchTerm), tc.lat, tc.lon)
				req, _ = http.NewRequest(tc.httpMethod, url, nil)
			)
			defer server.Close()
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("couldn't request: %v", err)
				t.FailNow()
			}
			if tc.expectedStatusCode != res.StatusCode {
				t.Errorf("expected status code %v but got %v", tc.expectedStatusCode, res.StatusCode)
			}
			defer res.Body.Close()
			if tc.expectedStatusCode != http.StatusOK {
				return
			}
			var actual = make([]item, 0)
			if err := json.NewDecoder(res.Body).Decode(&actual); err != nil {
				t.Errorf("couldn't read response payload into items: %v", err)
			}
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected %v but got %v", tc.expected, actual)
			}
		})
	}
}

func (db db) deleteIndex() {
	res1, err := db.client.DeleteIndex(db.index).Do(context.Background())
	if res1 == nil || !res1.Acknowledged {
		err = fmt.Errorf("db: DeleteIndex(%v) wasn't acknowledged by ES", db.index)
	}
	if err != nil {
		log.Printf("couldn't delete index %v\n", db.index)
	}
}

func randomHash() string {
	data := make([]byte, 10)
	rand.Read(data)
	return fmt.Sprintf("%x", sha256.Sum256(data))[0:6]
}
