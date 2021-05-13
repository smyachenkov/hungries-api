package main

type PlacesResponse struct {
	Places        []PlaceResponse `json:"places"`
	NextPageToken string          `json:"nextPageToken"`
}

type PlaceResponse struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
