package main

import (
	"net/http"
)

func main() {
	var db = mustNewDB("http://elasticsearch:9200", "elastic", "changeme", "item")
	db.mustReplaceIndex(mustReadCSVFromFile("dump.csv"))
	serve(&http.Server{Addr: ":8080", Handler: newEndpointHandler(db)})
}
