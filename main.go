package main

import (
	"flag"
	"net/http"
)

func main() {
	// For reviewing convenience, by default, this µs will load the sqlite csv dump onto a fresh `item` index.
	// This is not at all the responsibility of this µs.
	// A load balanced setup of replicas of this µs must always set the --no-replace-index flag.
	// In normal operation, one would expect a different process constantly populating the ES `item` index.
	var flagNoReplaceIndex = flag.Bool("no-replace-index", false, "whether to refresh the index on startup")
	flag.Parse()

	// Retries up to 10 times with 1 second delay while waiting for ES to become operational
	var db = mustNewDB("http://elasticsearch:9200", "elastic", "changeme", "item")

	if !*flagNoReplaceIndex {
		db.mustReplaceIndex(mustReadCSVFromFile("dump.csv"))
	}

	serve(&http.Server{Addr: ":8080", Handler: newEndpointHandler(db)})
}
