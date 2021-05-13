package main

type PlacesResponse struct {
	Places        []PlaceResponse `json:"places"`
	NextPageToken string          `json:"nextPageToken"`
}

type PlaceResponse struct {
	Name     string           `json:"name"`
	Url      string           `json:"url"`
	Location LocationResponse `json:"location"`
}

type LocationResponse struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}
