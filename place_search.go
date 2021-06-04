package main

import (
	"database/sql"
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"log"
	"math"
)

const MaxPhotoWidth = 600
const MaxPhotoHeight = 800

func FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string) (PlacesResponse, error) {
	nearbySearchResp, err := Dao.MapsApi.FindNearbyPlaces(coordinates, radius, pageToken)
	if err != nil {
		log.Print("Error finding places via Maps API " + err.Error())
		return PlacesResponse{}, err
	}

	var placesGoogleIds = make([]string, len(nearbySearchResp.Results))
	for i, googleId := range nearbySearchResp.Results {
		placesGoogleIds[i] = googleId.PlaceID
	}

	var places = make([]PlaceResponse, len(nearbySearchResp.Results))
	placesDb, _ := getPlaces(placesGoogleIds)

	for i, placeDb := range placesDb {
		placeResponse := PlaceResponse{
			Id:   placeDb.Id,
			Name: placeDb.Name,
			Url:  placeDb.Url,
			Location: LocationResponse{
				Latitude:  placeDb.Lat,
				Longitude: placeDb.Lng,
			},
			Distance: uint(getDistance(coordinates.Lat, coordinates.Lng, placeDb.Lat, placeDb.Lng)),
			PhotoUrl: placeDb.PhotoUrl.String,
		}
		places[i] = placeResponse
	}
	response := PlacesResponse{
		Places:        places,
		NextPageToken: nearbySearchResp.NextPageToken,
	}
	return response, nil
}

func getPlaces(googlePlaceId []string) ([]dao.PlaceDB, error) {
	// check db
	var places, e = Dao.PlacesDB.GetPlacesByPlaceIds(googlePlaceId)
	if e != nil {
		log.Print(e)
	}
	if len(places) == len(googlePlaceId) {
		return places, nil
	}

	// collect ids missing in db
	var missingPlacesGoogleIds []string
	for _, placeId := range googlePlaceId {
		if !contains(places, placeId) {
			missingPlacesGoogleIds = append(missingPlacesGoogleIds, placeId)
		}
	}

	// save new places
	for _, missingPlaceId := range missingPlacesGoogleIds {
		var placeDetailsResult, err = Dao.MapsApi.GetPlaceInfoFromMaps(missingPlaceId, []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskURL,
			maps.PlaceDetailsFieldMaskName,
			maps.PlaceDetailsFieldMaskGeometryLocationLat,
			maps.PlaceDetailsFieldMaskGeometryLocationLng,
			maps.PlaceDetailsFieldMaskPhotos,
		})
		if err != nil {
			log.Print(err)
			continue
		}
		// get photo and save it to cloud
		photoUrl, _ := uploadMainPhoto(missingPlaceId, placeDetailsResult.Photos)
		var newPlaceDb = dao.PlaceDB{
			GooglePlaceId: missingPlaceId,
			Name:          placeDetailsResult.Name,
			Url:           placeDetailsResult.URL,
			Lat:           placeDetailsResult.Geometry.Location.Lat,
			Lng:           placeDetailsResult.Geometry.Location.Lng,
			PhotoUrl:      sql.NullString{String: photoUrl, Valid: photoUrl != ""},
		}
		// todo batch save
		savedPlace, err := Dao.PlacesDB.SavePlace(newPlaceDb)
		if err != nil {
			log.Print(err)
			continue
		}
		places = append(places, *savedPlace)
	}
	return places, nil
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
