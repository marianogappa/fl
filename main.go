package main

import (
	"net/http"
)

func main() {
	var db = mustNewDB("http://127.0.0.1:9200", "elastic", "changeme", "item")
	db.mustReplaceIndex(mustReadCSVFromFile("dump.csv"))
	serve(&http.Server{Addr: ":8080", Handler: newEndpointHandler(db)})
}
