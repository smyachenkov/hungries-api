package dao

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type LikeDB struct {
	PlaceID  int32
	DeviceId string
	IsLiked  bool
}

type LikeDBService struct {
	DB *sql.DB
}

// SaveLike save new like or dislike for device
func (s *LikeDBService) SaveLike(deviceUUID string, placeID uint, isLiked bool) error {
	rows, err := s.DB.Query("insert into hungries.\"like\" (device_id, place_id, is_liked) "+
		"values ($1, $2, $3) "+
		"on conflict (device_id, place_id) do update set "+
		"update_date = now(), "+
		"is_liked    = excluded.is_liked",
		deviceUUID,
		placeID,
		isLiked,
	)
	defer rows.Close()
	if err != nil {
		log.Print("Error saving like for " + deviceUUID + " and place " + strconv.Itoa(int(placeID)))
		log.Print(err)
	}
	return err
}

// GetLikesForDevice get likes for device and internal places ids
func (s *LikeDBService) GetLikesForDevice(deviceUUID string, placeIds []uint) (map[uint]bool, error) {
	var result = make(map[uint]bool)
	var query = `select place_id, is_liked from hungries."like"
				 where device_id = $1 and place_id = any($2::int[])`

	var placesIdsString []string
	for _, p := range placeIds {
		placesIdsString = append(placesIdsString, fmt.Sprint(p))
	}
	var placeIdsParam = "{" + strings.Join(placesIdsString, ",") + "}"
	rows, err := s.DB.Query(query, deviceUUID, placeIdsParam)
	defer rows.Close()
	if err != nil {
		log.Print(fmt.Errorf("error getting likes for device %s and places %s", deviceUUID, placeIdsParam))
		return nil, err
	}
	for rows.Next() {
		var placeID uint
		var isLiked bool
		err := rows.Scan(&placeID, &isLiked)
		if err != nil {
			log.Print("Error reading like row")
			continue
		}
		result[placeID] = isLiked
	}
	return result, nil
}
