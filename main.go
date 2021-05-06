package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func findNearbyPlacesHandler(w http.ResponseWriter, r *http.Request) {
	pageTokenParam, hasToken := r.URL.Query()["pagetoken"]
	var pageToken string
	if hasToken {
		pageToken = pageTokenParam[0]
	}

	radius, _ := strconv.ParseUint(r.URL.Query()["radius"][0], 10, 64)

	coordinatesParam, hasCoordinates := r.URL.Query()["coordinates"]
	var coordinates maps.LatLng
	if hasCoordinates {
		latitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[0], 64)
		longitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[1], 64)
		coordinates = maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		}
	}

	places, err := FindNearbyPlaces(coordinates, uint(radius), pageToken)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Print("Error discovering places " + err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Missing $PORT environment variable")
	}
	router := mux.NewRouter()
	router.HandleFunc("/places", findNearbyPlacesHandler).Methods(http.MethodGet)
	http.ListenAndServe(":"+port, router)
}
