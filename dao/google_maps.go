package dao

import (
	"context"
	"errors"
	"googlemaps.github.io/maps"
)

type GoogleMapsAPIService struct {
	MapsClient *maps.Client
}

// GetPlaceInfoFromMaps get place info by google id
func (s *GoogleMapsAPIService) GetPlaceInfoFromMaps(placeId string, fields []maps.PlaceDetailsFieldMask) (maps.PlaceDetailsResult, error) {
	searchRequest := &maps.PlaceDetailsRequest{
		PlaceID: placeId,
		Fields:  fields,
	}
	detailsResp, err := s.MapsClient.PlaceDetails(context.Background(), searchRequest)
	if err != nil {
		return maps.PlaceDetailsResult{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	return detailsResp, nil
}

// FindNearbyPlaces find nearby places
func (s *GoogleMapsAPIService) FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string) (maps.PlacesSearchResponse, error) {
	searchRequest := &maps.NearbySearchRequest{
		Radius:    radius,
		PageToken: pageToken,
		Location:  &coordinates,
		Type:      maps.PlaceTypeRestaurant,
	}
	nearbySearchResp, err := s.MapsClient.NearbySearch(context.Background(), searchRequest)
	if err != nil {
		return maps.PlacesSearchResponse{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	return nearbySearchResp, nil
}
