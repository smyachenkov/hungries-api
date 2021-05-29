package dao

import (
	"database/sql"
	"log"
	"strconv"
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
	_, err := s.DB.Query("insert into hungries.\"like\" (device_id, place_id, is_liked) "+
		"values ($1, $2, $3) "+
		"on conflict (device_id, place_id) do update set "+
		"update_date = now(), "+
		"is_liked    = excluded.is_liked",
		deviceUUID,
		placeID,
		isLiked,
	)
	if err != nil {
		log.Print("Error saving like for " + deviceUUID + " and place " + strconv.Itoa(int(placeID)))
		log.Print(err)
	}
	return err
}
