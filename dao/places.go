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
		err := rows.Scan(
			&place.Id, &place.GooglePlaceId, &place.Name,
			&place.Url, &place.Lat, &place.Lng,
			&place.PhotoUrl,
		)
		if err != nil {
			log.Print("can't parse place")
		}
		result = append(result, place)
	}

	return result, nil
}

func (s *PlaceDbService) GetPlacesByPlaceIdsForDevice(googlePlaceIds []string) ([]PlaceDB, error) {
	var result []PlaceDB
	var query = `select ` + PlaceFields + `
				from hungries.place 
				where google_place_id = any($1::text[])`
	var placeIdsParam = "{" + strings.Join(googlePlaceIds, ",") + "}"
	rows, err := s.DB.Query(query, placeIdsParam)
	defer rows.Close()
	if err != nil {
		log.Print("Error searching places in db for places:  " + strings.Join(googlePlaceIds, " ") + " " + err.Error())
		return result, err
	}
	for rows.Next() {
		var place PlaceDB
		err := rows.Scan(
			&place.Id, &place.GooglePlaceId, &place.Name,
			&place.Url, &place.Lat, &place.Lng,
			&place.PhotoUrl,
		)
		if err != nil {
			log.Print("can't parse place")
			continue
		}
		result = append(result, place)
	}
	return result, nil
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

// SavePlaces save new places in batch
// todo add conflict check for same google id
func (s *PlaceDbService) SavePlaces(newPlaces []PlaceDB) []PlaceDB {
	const numberOfParams = 5
	var query = "insert into hungries.place (google_place_id, name, url, location, photo_url) values"
	var params = make([]interface{}, 0, len(newPlaces)*numberOfParams)
	for i, place := range newPlaces {
		query += fmt.Sprintf(
			"($%d, $%d, $%d, ST_GeomFromText($%d), nullif($%d, ''))",
			i*numberOfParams+1,
			i*numberOfParams+2,
			i*numberOfParams+3,
			i*numberOfParams+4,
			i*numberOfParams+5,
		)
		if i != len(newPlaces)-1 {
			query += ","
		}
		params = append(params,
			place.GooglePlaceId,
			place.Name,
			place.Url,
			LatLngToString(place.Lat, place.Lng),
			place.PhotoUrl,
		)
	}
	_, err := s.DB.Exec(query, params...)
	if err != nil {
		log.Print("error saving places: " + err.Error())
	}
	// return submitted places
	var placeGoogleIds []string
	for _, p := range newPlaces {
		placeGoogleIds = append(placeGoogleIds, p.GooglePlaceId)
	}
	result, _ := s.GetPlacesByPlaceIds(placeGoogleIds)
	return result
}

func LatLngToString(lat float64, lng float64) string {
	return fmt.Sprintf("Point(%f %f)", lat, lng)
}
