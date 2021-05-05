package main

type PlacesResponse struct {
	Places        []Place `json:"places"`
	NextPageToken string  `json:"nextPageToken"`
}

type Place struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
