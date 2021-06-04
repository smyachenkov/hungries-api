package dao

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type PlaceDB struct {
	Id            uint
	GooglePlaceId string
	Name          string
	Url           string
	Lat           float64
	Lng           float64
	PhotoUrl      sql.NullString
}

type PlaceDbService struct {
	DB *sql.DB
}

const PlaceFields = `id, google_place_id, name, url, ST_X(location::geometry), ST_Y(location::geometry), photo_url`

// PlaceExistsByGoogleId check if place exists by google id
func (s *PlaceDbService) PlaceExistsByGoogleId(googlePlaceId string) (bool, error) {
	var result bool
	row := s.DB.QueryRow(`select count(1) from hungries.place where google_place_id = $1`, googlePlaceId)
	err := row.Scan(&result)
	if err != nil {
		log.Print(err)
	}
	return result, nil
}

// PlaceExistsById check if place exists by internal id
func (s *PlaceDbService) PlaceExistsById(placeId uint) (bool, error) {
	var result bool
	row := s.DB.QueryRow(`select count(1) from hungries.place where id = $1`, placeId)
	err := row.Scan(&result)
	if err != nil {
		log.Print(err)
	}
	return result, nil
}

// GetPlaceByPlaceId get place buy it's googlePlaceId
func (s *PlaceDbService) GetPlaceByPlaceId(googlePlaceId string) (*PlaceDB, error) {
	var place PlaceDB
	row := s.DB.QueryRow(
		`select `+PlaceFields+` from hungries.place where google_place_id = $1`,
		googlePlaceId)
	err := row.Scan(&place.Id, &place.GooglePlaceId, &place.Name, &place.Url, &place.Lat, &place.Lng, &place.PhotoUrl)
	if err != nil {
		log.Print(err)
	}
	return &place, nil
}

func (s *PlaceDbService) GetPlacesByPlaceIds(googlePlaceIds []string) ([]PlaceDB, error) {
	var result []PlaceDB
	var query = `select ` + PlaceFields + ` from hungries.place where google_place_id = any($1::text[])`
	var param = "{" + strings.Join(googlePlaceIds, ",") + "}"
	rows, err := s.DB.Query(query, param)
	if err != nil {
		log.Print("Error searching places in db for places:  " + strings.Join(googlePlaceIds, " ") + " " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var place PlaceDB
		err := rows.Scan(&place.Id, &place.GooglePlaceId, &place.Name, &place.Url, &place.Lat, &place.Lng, &place.PhotoUrl)
		if err != nil {
			log.Print("can't parse place")
		}
		result = append(result, place)
	}

	return result, nil
}

// GetPlaceById get place buy it's id
func (s *PlaceDbService) GetPlaceById(id uint) (*PlaceDB, error) {
	var place PlaceDB
	row := s.DB.QueryRow(
		`select `+PlaceFields+` from hungries.place where id = $1`,
		id)
	err := row.Scan(&place.Id, &place.GooglePlaceId, &place.Name, &place.Url, &place.Lat, &place.Lng, &place.PhotoUrl)
	if err != nil {
		log.Print(err)
	}
	return &place, nil
}

// SavePlace save new place
func (s *PlaceDbService) SavePlace(newPlace PlaceDB) (*PlaceDB, error) {
	lastInsertId := uint(0)
	err := s.DB.QueryRow(`insert into
									hungries.place (google_place_id, name, url, location, photo_url) 
								values ($1, $2, $3, ST_GeomFromText($4), nullif($5, '')) returning id`,
		newPlace.GooglePlaceId,
		newPlace.Name,
		newPlace.Url,
		LatLngToString(newPlace.Lat, newPlace.Lng),
		newPlace.PhotoUrl,
	).Scan(&lastInsertId)
	if err != nil {
		log.Print(err)
	}
	return s.GetPlaceById(lastInsertId)
}

func LatLngToString(lat float64, lng float64) string {
	return fmt.Sprintf("Point(%f %f)", lat, lng)
}
