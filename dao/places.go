package dao

import (
	"database/sql"
	"log"
)

type PlaceDB struct {
	Id            int32
	GooglePlaceId string
	Name          string
	Url           string
}

type PlaceDbService struct {
	DB *sql.DB
}

// PlaceExists check if place exists buy it's google id
func (s *PlaceDbService) PlaceExists(googlePlaceId string) (bool, error) {
	var result bool
	row := s.DB.QueryRow(`select count(1) from hungries.place where google_place_id = $1`, googlePlaceId)
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result, nil
}

// GetPlaceByPlaceId get place buy it's googlePlaceId
func (s *PlaceDbService) GetPlaceByPlaceId(googlePlaceId string) (*PlaceDB, error) {
	var place PlaceDB
	row := s.DB.QueryRow(`select id, google_place_id, name, url from hungries.place where google_place_id = $1`, googlePlaceId)
	err := row.Scan(&place.Id, &place.GooglePlaceId, &place.Name, &place.Url)
	if err != nil {
		log.Fatal(err)
	}
	return &place, nil
}

// GetPlaceById get place buy it's id
func (s *PlaceDbService) GetPlaceById(id int32) (*PlaceDB, error) {
	var place PlaceDB
	row := s.DB.QueryRow(`select id, google_place_id, name, url from hungries.place where id = $1`, id)
	err := row.Scan(&place.Id, &place.GooglePlaceId, &place.Name, &place.Url)
	if err != nil {
		log.Fatal(err)
	}
	return &place, nil
}

// SavePlace save new place
func (s *PlaceDbService) SavePlace(newPlace PlaceDB) (*PlaceDB, error) {
	lastInsertId := 0
	err := s.DB.QueryRow(`insert into
    							hungries.place (google_place_id, name, url) 
								values ($1, $2, $3) returning id`,
		newPlace.GooglePlaceId,
		newPlace.Name,
		newPlace.Url,
	).Scan(&lastInsertId)
	if err != nil {
		log.Fatal(err)
	}
	return s.GetPlaceById(int32(lastInsertId))
}
