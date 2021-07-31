package main

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"math"
)

const MaxPhotoWidth = 600
const MaxPhotoHeight = 800

func FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string, deviceId string) (PlacesResponse, error) {
	log.WithFields(log.Fields{
		"coordinates": coordinates,
		"radius":      radius,
		"pageToken":   pageToken,
		"deviceId":    deviceId,
	}).Info("Searching places neardby")
	nearbySearchResp, err := Dao.MapsApi.FindNearbyPlaces(coordinates, radius, pageToken)
	if err != nil {
		log.WithField("error", err).Error("Error finding places via Maps API")
		return PlacesResponse{}, err
	}

	var placesGoogleIds = make([]string, len(nearbySearchResp.Results))
	for i, googleId := range nearbySearchResp.Results {
		placesGoogleIds[i] = googleId.PlaceID
	}

	// get places and likes info from db
	placesDb, _ := getPlaces(placesGoogleIds)

	var internalPlacesIds []uint
	for _, p := range placesDb {
		internalPlacesIds = append(internalPlacesIds, p.Id)
	}
	var likes map[uint]bool
	if deviceId != "" {
		likes, err = Dao.LikesDB.GetLikesForDevice(deviceId, internalPlacesIds)
	} else {
		likes = map[uint]bool{}
	}

	response := PlacesResponse{
		Places:        placeDBtoResponse(placesDb, likes, coordinates),
		NextPageToken: nearbySearchResp.NextPageToken,
	}
	return response, nil
}

func FindLikedPlaces(deviceId string, coordinates maps.LatLng) (PlacesResponse, error) {
	log.WithFields(log.Fields{
		"deviceId":    deviceId,
		"coordinates": coordinates,
	}).Info("Getting liked places")
	placesDb, err := Dao.PlacesDB.GetLikedPlacesForDevice(deviceId)
	if err != nil {
		return PlacesResponse{}, err
	}
	likes := make(map[uint]bool)
	for _, p := range placesDb {
		likes[p.Id] = true
	}
	response := PlacesResponse{
		Places: placeDBtoResponse(placesDb, likes, coordinates),
	}
	return response, nil
}

func placeDBtoResponse(placesDb []dao.PlaceDB, likes map[uint]bool, coordinates maps.LatLng) []PlaceResponse {
	var result []PlaceResponse
	for _, placeDb := range placesDb {
		var isLiked *bool
		if isLikedVal, ok := likes[placeDb.Id]; ok {
			isLiked = &isLikedVal
		}
		var photoUrl *string
		if placeDb.PhotoUrl.Valid {
			var photoUrlCopy = placeDb.PhotoUrl.String
			photoUrl = &photoUrlCopy
		}
		placeResponse := PlaceResponse{
			Id:   placeDb.Id,
			Name: placeDb.Name,
			Url:  placeDb.Url,
			Location: LocationResponse{
				Latitude:  placeDb.Lat,
				Longitude: placeDb.Lng,
			},
			Distance: uint(getDistance(coordinates.Lat, coordinates.Lng, placeDb.Lat, placeDb.Lng)),
			PhotoUrl: photoUrl,
			IsLiked:  isLiked,
		}
		result = append(result, placeResponse)
	}
	return result
}

func getPlaces(googlePlaceIds []string) ([]dao.PlaceDB, error) {
	// check db
	var existingPlaces, err = Dao.PlacesDB.GetPlacesByPlaceIdsForDevice(googlePlaceIds)
	if err != nil {
		log.WithFields(log.Fields{
			"error":          err,
			"googlePlaceIds": googlePlaceIds,
		}).Info("Error getting places from db")
	}
	if len(existingPlaces) == len(googlePlaceIds) {
		return existingPlaces, nil
	}

	var result []dao.PlaceDB
	for _, p := range existingPlaces {
		result = append(result, p)
	}

	// collect ids absent in db
	var missingPlacesGoogleIds []string
	for _, placeId := range googlePlaceIds {
		if !contains(existingPlaces, placeId) {
			missingPlacesGoogleIds = append(missingPlacesGoogleIds, placeId)
		}
	}

	// get new places from google maps API
	newPlacesToSaveChan := make(chan dao.PlaceDB)
	for _, missingPlaceId := range missingPlacesGoogleIds {
		go getInfoAndUploadPhoto(missingPlaceId, newPlacesToSaveChan)
	}

	var newPlacesToSave []dao.PlaceDB
	for i := 0; i < len(missingPlacesGoogleIds); i++ {
		place := <-newPlacesToSaveChan
		newPlacesToSave = append(newPlacesToSave, place)
	}

	// save new places
	var newSavedPlaces = Dao.PlacesDB.SavePlaces(newPlacesToSave)
	for _, p := range newSavedPlaces {
		result = append(result, p)
	}
	return result, nil
}

func getInfoAndUploadPhoto(googlePlaceID string, newPlacesToSave chan dao.PlaceDB) {
	var placeDetailsResult, err = Dao.MapsApi.GetPlaceInfoFromMaps(googlePlaceID, []maps.PlaceDetailsFieldMask{
		maps.PlaceDetailsFieldMaskURL,
		maps.PlaceDetailsFieldMaskName,
		maps.PlaceDetailsFieldMaskGeometryLocationLat,
		maps.PlaceDetailsFieldMaskGeometryLocationLng,
		maps.PlaceDetailsFieldMaskPhotos,
	})
	if err != nil {
		log.Print(err)
		return
	}
	// get photo and save it to cloud
	photoUrl, _ := uploadMainPhoto(googlePlaceID, placeDetailsResult.Photos)
	var newPlaceDb = dao.PlaceDB{
		GooglePlaceId: googlePlaceID,
		Name:          placeDetailsResult.Name,
		Url:           placeDetailsResult.URL,
		Lat:           placeDetailsResult.Geometry.Location.Lat,
		Lng:           placeDetailsResult.Geometry.Location.Lng,
		PhotoUrl:      sql.NullString{String: photoUrl, Valid: photoUrl != ""},
	}
	newPlacesToSave <- newPlaceDb
}

func contains(s []dao.PlaceDB, e string) bool {
	for _, a := range s {
		if a.GooglePlaceId == e {
			return true
		}
	}
	return false
}

func uploadMainPhoto(placeId string, photos []maps.Photo) (string, error) {
	if len(photos) == 0 {
		return "", nil
	}
	firstPhoto := photos[0]
	photoReference := firstPhoto.PhotoReference
	photo, err := Dao.MapsApi.GetPhoto(photoReference, MaxPhotoWidth, MaxPhotoHeight)
	if err != nil {
		log.Print(err)
		return "", err
	}
	photoUrl, err := Dao.CloudStorage.UploadPhoto(placeId, photo.Data)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return photoUrl, nil
}

// Distance between 2 points in meters
// See https://gist.github.com/cdipaolo/d3f8db3848278b49db68
// http://en.wikipedia.org/wiki/Haversine_formula
func getDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
