package main

import (
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"log"
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
