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
	"os"
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
			useCSVItems        bool
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
				name:               "incorrect latitude returns Bad Request",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "camera",
				lat:                "not a lat",
				lon:                "0",
				expected:           []item{},
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				name:               "empty longitude returns Bad Request",
				items:              `"camera",51,0,london/camera,[]`,
				httpMethod:         "GET",
				endpoint:           "/search",
				searchTerm:         "camera",
				lat:                "51",
				lon:                "",
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
			{
				name:        "returns up to 20 entries",
				useCSVItems: true,
				httpMethod:  "GET",
				endpoint:    "/search",
				searchTerm:  "cameras",
				lat:         "51.4",
				lon:         "-0.1",
				expected: []item{
					{Name: "Panasonic GH5 Camera with Vlog", Location: location{Lat: 51.4177208, Lon: -0.122357696}, URL: "london/hire-panasonic-gh5-camera--28584820", ImgURLs: []string{"panasonic-gh5-camera--49860290.jpg", "panasonic-gh5-camera--01111925.jpg"}},
					{Name: "Canon 7D Camera", Location: location{Lat: 51.4389496, Lon: -0.154008105}, URL: "london/hire-canon-7d-camera-11908390", ImgURLs: []string{"canon-7d-camera-45742621.jpg", "canon-7d-camera-71254330.jpg"}},
					{Name: "Manfrotto Fluid Video / Camera Monopod with Head", Location: location{Lat: 51.406601, Lon: -0.178710699}, URL: "london/hire-manfrotto-fluid-video--camera-monopod-with-head-38287382", ImgURLs: []string{"manfrotto-fluid-video--camera-monopod-with-head-87071795.jpg", "manfrotto-fluid-video--camera-monopod-with-head-94628553.jpg", "manfrotto-fluid-video--camera-monopod-with-head-96940506.jpg", "manfrotto-fluid-video--camera-monopod-with-head-52214974.jpg", "manfrotto-fluid-video--camera-monopod-with-head-11538367.jpg", "manfrotto-fluid-video--camera-monopod-with-head-64470847.jpg"}},
					{Name: "Canon 5DSR Digital Camera", Location: location{Lat: 51.4122009, Lon: -0.0116203995}, URL: "london/hire-canon-5dsr-digital-camera-52417627", ImgURLs: []string{"canon-5dsr-digital-camera-96251280.jpg", "canon-5dsr-digital-camera-24006311.jpg"}},
					{Name: "Head Strap (GoPro Camera Accessory)", Location: location{Lat: 51.4070053, Lon: -0.177427202}, URL: "london/hire-head-strap-gopro-camera-accessory-76409921", ImgURLs: []string{"head-strap-gopro-camera-accessory-52110933.jpg", "head-strap-gopro-camera-accessory-39408208.jpg", "head-strap-gopro-camera-accessory-65384018.jpg"}},
					{Name: "GoPro Camera Hero 3+ Black & Accessories", Location: location{Lat: 51.4071999, Lon: -0.178354993}, URL: "london/hire-gopro-camera-hero-3-black--accessories-41006356", ImgURLs: []string{"gopro-camera-hero-3-black--accessories-23523806.JPG", "gopro-camera-hero-3-black--accessories-51437059.JPG", "gopro-camera-hero-3-black--accessories-97436384.JPG", "gopro-camera-hero-3-black--accessories-99703879.JPG", "gopro-camera-hero-3-black--accessories-77474823.JPG", "gopro-camera-hero-3-black--accessories-25426488.JPG", "gopro-camera-hero-3-black--accessories-03322570.JPG", "gopro-camera-hero-3-black--accessories-70811534.JPG"}},
					{Name: "Manfrotto Fluid Video / Camera Monopod with Head", Location: location{Lat: 51.4071884, Lon: -0.177306801}, URL: "london/hire-manfrotto-fluid-video--camera-monopod-with-head-37416021", ImgURLs: []string{"manfrotto-fluid-video--camera-monopod-with-head-49214569.jpg", "manfrotto-fluid-video--camera-monopod-with-head-53827599.jpg", "manfrotto-fluid-video--camera-monopod-with-head-69895924.jpg", "manfrotto-fluid-video--camera-monopod-with-head-30738963.jpg", "manfrotto-fluid-video--camera-monopod-with-head-99815617.jpg", "manfrotto-fluid-video--camera-monopod-with-head-23188000.jpg"}},
					{Name: "Canon Digital SLR Camera EOS 5D Mark II", Location: location{Lat: 51.4062958, Lon: -0.178215206}, URL: "london/hire-canon-digital-slr-camera-eos-5d-mark-ii-35407495", ImgURLs: []string{"canon-digital-slr-camera-eos-5d-mark-ii-60186903.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-25651637.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-53079862.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-86083773.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-16534503.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-38237692.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-17686152.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-52294656.jpg"}},
					{Name: "Canon Speedlite 580EX // Shoe Mount Camera Flash", Location: location{Lat: 51.4077187, Lon: -0.178006798}, URL: "london/hire-canon-speedlite-580ex--shoe-mount-camera-flash-31715615", ImgURLs: []string{"canon-speedlite-580ex--shoe-mount-camera-flash-70845653.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-74719760.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-57384401.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-10008141.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-31921628.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-20426928.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-05713437.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-59981253.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-77703997.JPG", "canon-speedlite-580ex--shoe-mount-camera-flash-81121627.JPG"}},
					{Name: "Pair of Manfrotto Fluid Video / Camera Monopods with Head", Location: location{Lat: 51.4076614, Lon: -0.177564099}, URL: "london/hire-pair-of-manfrotto-fluid-video--camera-monopods-with-head-15731526", ImgURLs: []string{"pair-of-manfrotto-fluid-video--camera-monopods-with-head-80097844.jpg", "pair-of-manfrotto-fluid-video--camera-monopods-with-head-49475094.jpg", "pair-of-manfrotto-fluid-video--camera-monopods-with-head-81249926.jpg", "pair-of-manfrotto-fluid-video--camera-monopods-with-head-05354735.jpg", "pair-of-manfrotto-fluid-video--camera-monopods-with-head-92845802.jpg", "pair-of-manfrotto-fluid-video--camera-monopods-with-head-05502261.jpg"}},
					{Name: "C100 body only in PELI case   cr", Location: location{Lat: 51.3759995, Lon: -0.0788893998}, URL: "london/hire-canon-c100-cinema-camera-body-only-16569490", ImgURLs: []string{"canon-c100-cinema-camera-body-only-66523087.JPG", "canon-c100-cinema-camera-body-only-46356872.JPG", "canon-c100-cinema-camera-body-only-53450635.JPG"}},
					{Name: "Canon EF 16-35mm f/2.8L II USM Camera Lens", Location: location{Lat: 51.4070206, Lon: -0.178750798}, URL: "london/hire-canon-ef-1635mm-f28l-ii-usm-camera-lens-40663112", ImgURLs: []string{"canon-ef-1635mm-f28l-ii-usm-camera-lens-66951881.JPG", "canon-ef-1635mm-f28l-ii-usm-camera-lens-68607067.JPG", "canon-ef-1635mm-f28l-ii-usm-camera-lens-98341001.JPG", "canon-ef-1635mm-f28l-ii-usm-camera-lens-01771477.JPG", "canon-ef-1635mm-f28l-ii-usm-camera-lens-69149056.JPG"}},
					{Name: "POV Pole 36\" // Extendable Selfie Stick for GoPro Camera", Location: location{Lat: 51.4078369, Lon: -0.178008795}, URL: "london/hire-pov-pole-36--extendable-selfie-stick-for-gopro-camera-07659865", ImgURLs: []string{"pov-pole-36--extendable-selfie-stick-for-gopro-camera-93730626.jpg", "pov-pole-36--extendable-selfie-stick-for-gopro-camera-41581058.jpg", "pov-pole-36--extendable-selfie-stick-for-gopro-camera-33876904.jpg"}},
					{Name: "Chesty // Chest Mount Harness (GoPro Camera Accessory)", Location: location{Lat: 51.4076424, Lon: -0.177775696}, URL: "london/hire-gopro-chesty--chest-mount-harness-accessory-36614055", ImgURLs: []string{"chesty--chest-mount-harness-gopro-camera-accessory-70592595.jpg", "chesty--chest-mount-harness-gopro-camera-accessory-55849160.jpg"}},
					{Name: "Large Format 6x7 Linhof Technikardan film camera & lens package", Location: location{Lat: 51.4128189, Lon: -0.0112944003}, URL: "london/hire-6x7-linhof-technikardan-film-camera--lens-package-31938782", ImgURLs: []string{"6x7-linhof-technikardan-large-format-film-camera--lens-package-68318076.jpg", "6x7-linhof-technikardan-large-format-film-camera--lens-package-10536001.jpg", "6x7-linhof-technikardan-large-format-film-camera--lens-package-83573667.jpg", "6x7-linhof-technikardan-large-format-film-camera--lens-package-90609664.jpg", "6x7-linhof-technikardan-large-format-film-camera--lens-package-64305866.jpg"}},
					{Name: "Canon Digital SLR Camera EOS 5D Mark II", Location: location{Lat: 51.4067993, Lon: -0.178405598}, URL: "london/hire-canon-digital-slr-camera-eos-5d-mark-ii-76491193", ImgURLs: []string{"canon-digital-slr-camera-eos-5d-mark-ii-96027785.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-60984855.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-89961555.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-53567215.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-42246715.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-76812183.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-86511701.jpg", "canon-digital-slr-camera-eos-5d-mark-ii-95951704.jpg"}},
					{Name: "Canon EF 85mm f/1.2L II USM Prime Camera Lens", Location: location{Lat: 51.4067802, Lon: -0.178617507}, URL: "london/hire-canon-ef-85mm-f12l-ii-usm-lens-02490552", ImgURLs: []string{"canon-ef-85mm-f12l-ii-usm-camera-lens-95103363.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-48829504.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-30872497.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-90499293.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-54066280.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-52492167.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-04597700.JPG", "canon-ef-85mm-f12l-ii-usm-camera-lens-68629883.JPG"}},
					{Name: "Universal Camera shoulder rig with Mattebox and handles", Location: location{Lat: 51.4744186, Lon: -0.0438758992}, URL: "london/hire-universal-camera-shoulder-rig-with-mattebox-and-handles-68680281", ImgURLs: []string{"universal-camera-shoulder-rig-with-mattebox-and-handles-77739438.JPG", "universal-camera-shoulder-rig-with-mattebox-and-handles-72790309.JPG", "universal-camera-shoulder-rig-with-mattebox-and-handles-07359085.JPG", "universal-camera-shoulder-rig-with-mattebox-and-handles-51473279.JPG", "universal-camera-shoulder-rig-with-mattebox-and-handles-01596921.png", "universal-camera-shoulder-rig-with-mattebox-and-handles-96785278.png", "universal-camera-shoulder-rig-with-mattebox-and-handles-93148160.jpg", "universal-camera-shoulder-rig-with-mattebox-and-handles-25241240.jpg", "universal-camera-shoulder-rig-with-mattebox-and-handles-16453768.png", "universal-camera-shoulder-rig-with-mattebox-and-handles-17952315.png", "universal-camera-shoulder-rig-with-mattebox-and-handles-47829218.png"}},
					{Name: "Canon 24-70mm F/2.8 L USM Macro EF Mount Camera Lens", Location: location{Lat: 51.4071808, Lon: -0.179074407}, URL: "london/hire-canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-64216119", ImgURLs: []string{"canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-99693795.JPG", "canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-58741569.JPG", "canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-82391349.JPG", "canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-66301995.JPG", "canon-2470mm-f28-l-usm-macro-ef-mount-camera-lens-93642557.JPG"}},
					{Name: "Canon EOS 6D Camera", Location: location{Lat: 51.4658012, Lon: -0.0261311997}, URL: "london/hire-canon-eos-6d-camera-27115255", ImgURLs: []string{"canon-eos-6d-camera-54715632.jpg", "canon-eos-6d-camera-49293873.jpg"}},
				},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:        "Finds 3 results for Cort Bass; most relevant by searchTerm and location first",
				useCSVItems: true,
				httpMethod:  "GET",
				endpoint:    "/search",
				searchTerm:  "Cort Bass",
				lat:         "51.4",
				lon:         "-0.1", // <-- relevant parameter
				expected: []item{
					{Name: "Cort Acoustic Bass guitar", Location: location{Lat: 51.5711136, Lon: -0.123528004}, URL: "london/hire-cort-acoustic-bass-guitar-07529191", ImgURLs: []string{"cort-acoustic-bass-guitar-81141134.jpg"}},
					{Name: "Fender Jazz Bass American ", Location: location{Lat: 51.5301895, Lon: 0.0407329984}, URL: "london/hire-fender-jazz-bass-american--23230868", ImgURLs: []string{"fender-jazz-bass-american--99722657.JPG", "fender-jazz-bass-american--26390453.JPG", "fender-jazz-bass-american--26488256.JPG"}},
					{Name: "Novation Bass Station II", Location: location{Lat: 51.5551682, Lon: -0.207050607}, URL: "london/hire-novation-bass-station-ii-57382981", ImgURLs: []string{"novation-bass-station-ii-32986505.jpg"}},
				},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:        "Same Cort Bass search, but far enough from closest searchTerm match that it becomes 2nd",
				useCSVItems: true,
				httpMethod:  "GET",
				endpoint:    "/search",
				searchTerm:  "Cort Bass",
				lat:         "51.4",
				lon:         "0.2", // <-- relevant parameter changed
				expected: []item{
					{Name: "Fender Jazz Bass American ", Location: location{Lat: 51.5301895, Lon: 0.0407329984}, URL: "london/hire-fender-jazz-bass-american--23230868", ImgURLs: []string{"fender-jazz-bass-american--99722657.JPG", "fender-jazz-bass-american--26390453.JPG", "fender-jazz-bass-american--26488256.JPG"}},
					{Name: "Cort Acoustic Bass guitar", Location: location{Lat: 51.5711136, Lon: -0.123528004}, URL: "london/hire-cort-acoustic-bass-guitar-07529191", ImgURLs: []string{"cort-acoustic-bass-guitar-81141134.jpg"}},
					{Name: "Novation Bass Station II", Location: location{Lat: 51.5551682, Lon: -0.207050607}, URL: "london/hire-novation-bass-station-ii-57382981", ImgURLs: []string{"novation-bass-station-ii-32986505.jpg"}},
				},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:        "Geogr. closest to Novation Bass but still within decay range to closest searchTerm match",
				useCSVItems: true,
				httpMethod:  "GET",
				endpoint:    "/search",
				searchTerm:  "Cort Bass",
				lat:         "51.4",
				lon:         "-0.2", // <-- relevant parameter changed
				expected: []item{
					{Name: "Cort Acoustic Bass guitar", Location: location{Lat: 51.5711136, Lon: -0.123528004}, URL: "london/hire-cort-acoustic-bass-guitar-07529191", ImgURLs: []string{"cort-acoustic-bass-guitar-81141134.jpg"}},
					{Name: "Novation Bass Station II", Location: location{Lat: 51.5551682, Lon: -0.207050607}, URL: "london/hire-novation-bass-station-ii-57382981", ImgURLs: []string{"novation-bass-station-ii-32986505.jpg"}},
					{Name: "Fender Jazz Bass American ", Location: location{Lat: 51.5301895, Lon: 0.0407329984}, URL: "london/hire-fender-jazz-bass-american--23230868", ImgURLs: []string{"fender-jazz-bass-american--99722657.JPG", "fender-jazz-bass-american--26390453.JPG", "fender-jazz-bass-american--26488256.JPG"}},
				},
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
			db.index = "test_items_" + randomHash()
			loadItemsIntoTestIndex(tc.items, tc.useCSVItems, db, t)
			defer db.deleteIndex()
			actualItems, actualStatusCode := testRequest(tc.httpMethod, tc.endpoint, tc.searchTerm, tc.lat, tc.lon, db, t)
			if tc.expectedStatusCode != actualStatusCode {
				t.Errorf("expected status code %v but got %v", tc.expectedStatusCode, actualStatusCode)
				t.FailNow()
			}
			if !reflect.DeepEqual(tc.expected, actualItems) {
				t.Errorf("expected %v but got %#v", tc.expected, actualItems)
			}
		})
	}
}

func loadItemsIntoTestIndex(strItems string, useCSVItems bool, db db, t *testing.T) {
	items, err := readCSV(strings.NewReader(strItems))
	if useCSVItems {
		var fh *os.File
		fh, err = os.Open("dump.csv")
		if err == nil {
			items, err = readCSV(fh)
		}
	}
	if err != nil {
		t.Errorf("couldn't read items: %v", err)
		t.FailNow()
	}
	if err := db.replaceIndex(items); err != nil {
		t.Errorf("couldn't replace index: %v", err)
		t.FailNow()
	}
}

func testRequest(httpMethod, endpoint, searchTerm, lat, lon string, db db, t *testing.T) ([]item, int) {
	var (
		server = httptest.NewServer(http.HandlerFunc(newEndpointHandler(db).ServeHTTP))
		client = http.Client{}
		url    = fmt.Sprintf("%v%v?searchTerm=%v&lat=%v&lng=%v",
			server.URL, endpoint, url.PathEscape(searchTerm), lat, lon)
		req, _ = http.NewRequest(httpMethod, url, nil)
	)
	defer server.Close()
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't request: %v", err)
		t.FailNow()
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return []item{}, res.StatusCode
	}
	var actualItems = make([]item, 0)
	if err := json.NewDecoder(res.Body).Decode(&actualItems); err != nil {
		t.Errorf("couldn't read response payload into items: %v", err)
		t.FailNow()
	}
	return actualItems, res.StatusCode
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
