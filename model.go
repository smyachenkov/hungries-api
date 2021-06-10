package main

type PlacesResponse struct {
	Places        []PlaceResponse `json:"places"`
	NextPageToken string          `json:"nextPageToken"`
}

type PlaceResponse struct {
	Id       uint             `json:"id"`
	Name     string           `json:"name"`
	Url      string           `json:"url"`
	Location LocationResponse `json:"location"`
	Distance uint             `json:"distance"`
	PhotoUrl string           `json:"photoUrl"`
	IsLiked  *bool            `json:"isLiked"`
}

type LocationResponse struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}
