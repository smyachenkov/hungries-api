package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func findNearbyPlacesHandler(w http.ResponseWriter, r *http.Request) {
	deviceId := getStringParamWithDefault(r.URL.Query(), "device", "")
	pageToken, err := getStringParamRequired(r.URL.Query(), "pagetoken")
	radius, _ := strconv.ParseUint(r.URL.Query()["radius"][0], 10, 64)
	coordinates, err := getCoordinatesParam(r.URL.Query(), "coordinates")
	if err != nil {
		log.WithField("error", err).Error("Incorrect input data")
		return
	}

	places, err := FindNearbyPlaces(coordinates, uint(radius), pageToken, deviceId)
	if err != nil {
		log.WithField("error", err).Error("Error discovering places")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func getLikedPlacesHandler(w http.ResponseWriter, r *http.Request) {
	deviceId, err := getStringParamRequired(r.URL.Query(), "device")
	coordinates, err := getCoordinatesParam(r.URL.Query(), "coordinates")
	if err != nil {
		log.WithField("error", err).Error("Incorrect input data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	places, err := FindLikedPlaces(deviceId, coordinates)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func saveLikeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	placeId, err := strconv.ParseUint(vars["place"], 10, 64)
	deviceUUID := vars["device"]
	isLiked, err := strconv.ParseBool(vars["liked"])

	if err != nil {
		log.WithField("error", err).Error("Incorrect input data")
	}

	// check if place exist
	placeExist, err := Dao.PlacesDB.PlaceExistsById(uint(placeId))
	if err != nil {
		log.WithField("error", err).Error("Can't find place")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !placeExist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = Dao.LikesDB.SaveLike(deviceUUID, uint(placeId), isLiked)
	if err != nil {
		log.WithField("error", err).Error("Error submitting like/dislike")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func getStringParamRequired(values url.Values, paramName string) (string, error) {
	paramValues, hasParam := values[paramName]
	var value string
	if hasParam {
		value = paramValues[0]
	} else {
		return "", errors.New("missing param " + paramName)
	}
	return value, nil
}

func getStringParamWithDefault(values url.Values, paramName string, defaultValue string) string {
	paramValues, hasParam := values[paramName]
	var value string
	if hasParam {
		value = paramValues[0]
	} else {
		return defaultValue
	}
	return value
}

func getCoordinatesParam(values url.Values, paramName string) (maps.LatLng, error) {
	coordinatesParam, hasCoordinates := values[paramName]
	var coordinates maps.LatLng
	if hasCoordinates {
		latitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[0], 64)
		longitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[1], 64)
		coordinates = maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		}
	}
	return coordinates, nil
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
