package main

import (
	"crypto/subtle"
	"encoding/json"
	"github.com/gorilla/mux"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
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

func saveLikeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	placeId, _ := strconv.ParseUint(vars["place"], 10, 64)
	deviceUUID := vars["device"]
	isLiked, _ := strconv.ParseBool(vars["liked"])

	// check if place exist
	placeExist, err := Dao.PlacesDB.PlaceExistsById(uint(placeId))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !placeExist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = Dao.LikesDB.SaveLike(deviceUUID, uint(placeId), isLiked)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("Error submitting like " + err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// BasicAuth basic auth wrapper for handlers, see https://stackoverflow.com/questions/21936332/
func BasicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Hungries API"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
