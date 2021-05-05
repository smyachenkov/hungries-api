package main

import (
	"context"
	"errors"
	"googlemaps.github.io/maps"
	"os"
)

var GoogleMapsApiKey = os.Getenv("GOOGLE_MAPS_API_KEY")

var mapsClient, _ = maps.NewClient(maps.WithAPIKey(GoogleMapsApiKey))

func FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string) (PlacesResponse, error) {
	searchRequest := &maps.NearbySearchRequest{
		Radius:    radius,
		PageToken: pageToken,
		Location:  &coordinates,
		Type:      maps.PlaceTypeRestaurant,
	}
	nearbySearchResp, err := mapsClient.NearbySearch(context.Background(), searchRequest)
	if err != nil {
		return PlacesResponse{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	var places = make([]Place, len(nearbySearchResp.Results))
	for i := 0; i < len(nearbySearchResp.Results); i++ {
		currentPlace := nearbySearchResp.Results[i]
		placeInfo, _ := getPlaceInfo(currentPlace.PlaceID, []maps.PlaceDetailsFieldMask{maps.PlaceDetailsFieldMaskURL})
		place := Place{
			Name: currentPlace.Name,
			Url:  placeInfo.URL,
		}
		places[i] = place
	}
	response := PlacesResponse{
		Places:        places,
		NextPageToken: nearbySearchResp.NextPageToken,
	}
	return response, nil
}

func getPlaceInfo(placeId string, fields []maps.PlaceDetailsFieldMask) (maps.PlaceDetailsResult, error) {
	searchRequest := &maps.PlaceDetailsRequest{
		PlaceID: placeId,
		Fields:  fields,
	}
	detailsResp, err := mapsClient.PlaceDetails(context.Background(), searchRequest)
	if err != nil {
		return maps.PlaceDetailsResult{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	return detailsResp, nil
}
