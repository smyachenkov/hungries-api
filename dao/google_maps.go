package dao

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

type GoogleMapsAPIService struct {
	MapsClient *maps.Client
}

// GetPlaceInfoFromMaps get place info by google id
func (s *GoogleMapsAPIService) GetPlaceInfoFromMaps(placeId string, fields []maps.PlaceDetailsFieldMask) (maps.PlaceDetailsResult, error) {
	log.WithField("placeId", placeId).Info("Getting info for new place from Google Maps API")
	searchRequest := &maps.PlaceDetailsRequest{
		PlaceID: placeId,
		Fields:  fields,
	}
	detailsResp, err := s.MapsClient.PlaceDetails(context.Background(), searchRequest)
	if err != nil {
		log.WithField("placeId", placeId).Error("Failed to get place details from Google Maps API")
		return maps.PlaceDetailsResult{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	return detailsResp, nil
}

// FindNearbyPlaces find nearby places
func (s *GoogleMapsAPIService) FindNearbyPlaces(coordinates maps.LatLng, radius uint, pageToken string) (maps.PlacesSearchResponse, error) {
	log.WithFields(log.Fields{
		"coordinates": coordinates,
		"radius":      radius,
	}).Info("Searching for nearby places via Google Maps API")
	searchRequest := &maps.NearbySearchRequest{
		Radius:    radius,
		PageToken: pageToken,
		Location:  &coordinates,
		Type:      maps.PlaceTypeRestaurant,
	}
	nearbySearchResp, err := s.MapsClient.NearbySearch(context.Background(), searchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"coordinates": coordinates,
			"radius":      radius,
		}).Error("Failed to get nearby places from Google Maps API")
		return maps.PlacesSearchResponse{}, errors.New("Error requesting Maps API: " + err.Error())
	}
	return nearbySearchResp, nil
}

// GetPhoto get photo of the place
func (s *GoogleMapsAPIService) GetPhoto(photoReference string, width uint, height uint) (maps.PlacePhotoResponse, error) {
	log.WithField("photoReference", photoReference).Info("Getting photo from Google Maps API")
	photoRequest := &maps.PlacePhotoRequest{
		PhotoReference: photoReference,
		MaxHeight:      height,
		MaxWidth:       width,
	}
	placePhotoResponse, err := s.MapsClient.PlacePhoto(context.Background(), photoRequest)
	if err != nil {
		log.WithField("photoReference", photoReference).Error("Failed to get photo from Google Maps API")
		return maps.PlacePhotoResponse{}, err
	}
	return placePhotoResponse, nil
}
