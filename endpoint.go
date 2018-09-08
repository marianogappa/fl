package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type endpointHandler struct {
	db db
}

func newEndpointHandler(db db) endpointHandler {
	return endpointHandler{db}
}

func (eh endpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/search" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var err error
	searchTerm := r.URL.Query().Get("searchTerm")
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	items, err := eh.db.search(searchTerm, location{Lat: lat, Lon: lng})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
