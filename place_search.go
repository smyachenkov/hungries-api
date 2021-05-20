package main

import (
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"log"
	"math"
)

func FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string) (PlacesResponse, error) {
	nearbySearchResp, err := Dao.MapsApi.FindNearbyPlaces(coordinates, radius, pageToken)
	if err != nil {
		log.Print("Error finding places via Maps API " + err.Error())
		return PlacesResponse{}, err
	}
	var places = make([]PlaceResponse, len(nearbySearchResp.Results))
	for i := 0; i < len(nearbySearchResp.Results); i++ {
		currentPlace := nearbySearchResp.Results[i]
		placeInfo, _ := getPlace(currentPlace.PlaceID)
		place := PlaceResponse{
			Name: placeInfo.Name,
			Url:  placeInfo.Url,
			Location: LocationResponse{
				Latitude:  placeInfo.Lat,
				Longitude: placeInfo.Lng,
			},
			Distance: int32(getDistance(coordinates.Lat, coordinates.Lng, placeInfo.Lat, placeInfo.Lng)),
		}
		places[i] = place
	}
	response := PlacesResponse{
		Places:        places,
		NextPageToken: nearbySearchResp.NextPageToken,
	}
	return response, nil
}

func getPlace(placeId string) (dao.PlaceDB, error) {
	// check db
	exists, _ := Dao.PlacesDB.PlaceExists(placeId)
	if exists {
		var place, err = Dao.PlacesDB.GetPlaceByPlaceId(placeId)
		if err != nil {
			log.Print(err)
			return dao.PlaceDB{}, err
		}
		return *place, err
	}
	// not in db - get from maps API
	var placeDetailsResult, err = Dao.MapsApi.GetPlaceInfoFromMaps(placeId, []maps.PlaceDetailsFieldMask{
		maps.PlaceDetailsFieldMaskURL,
		maps.PlaceDetailsFieldMaskName,
		maps.PlaceDetailsFieldMaskGeometryLocationLat,
		maps.PlaceDetailsFieldMaskGeometryLocationLng,
	})
	if err != nil {
		log.Print(err)
		return dao.PlaceDB{}, err
	}
	// save
	var newPlaceDb = dao.PlaceDB{
		GooglePlaceId: placeId,
		Name:          placeDetailsResult.Name,
		Url:           placeDetailsResult.URL,
		Lat:           placeDetailsResult.Geometry.Location.Lat,
		Lng:           placeDetailsResult.Geometry.Location.Lng,
	}
	result, err := Dao.PlacesDB.SavePlace(newPlaceDb)
	if err != nil {
		log.Print(err)
		return dao.PlaceDB{}, err
	}
	return *result, err
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
