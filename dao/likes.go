package dao

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type LikeDB struct {
	PlaceID  int32
	DeviceId string
	IsLiked  bool
}

type LikeDBService struct {
	DB *sql.DB
}

// SaveLike save new like or dislike for userId
func (s *LikeDBService) SaveLike(userId string, placeID uint, isLiked bool) error {
	log.WithFields(log.Fields{
		"userId":  userId,
		"place":   strconv.Itoa(int(placeID)),
		"isLiked": isLiked,
	}).Info("Saving like")
	rows, err := s.DB.Query("insert into hungries.\"like\" (user_id, place_id, is_liked) "+
		"values ($1, $2, $3) "+
		"on conflict (user_id, place_id) do update set "+
		"update_date = now(), "+
		"is_liked    = excluded.is_liked",
		userId,
		placeID,
		isLiked,
	)
	defer rows.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"userId":  userId,
			"place":   strconv.Itoa(int(placeID)),
			"isLiked": isLiked,
			"error":   err,
		}).Error("Error saving like")
	}
	return err
}

// GetLikesForDevice get likes for device and internal places ids
func (s *LikeDBService) GetLikesForDevice(userId string, placeIds []uint) (map[uint]bool, error) {
	log.WithField("userId", userId).Info("Getting likes for userId")
	var result = make(map[uint]bool)
	var query = `select place_id, is_liked from hungries."like"
				 where user_id = $1 and place_id = any($2::int[])`

	var placesIdsString []string
	for _, p := range placeIds {
		placesIdsString = append(placesIdsString, fmt.Sprint(p))
	}
	var placeIdsParam = "{" + strings.Join(placesIdsString, ",") + "}"
	rows, err := s.DB.Query(query, userId, placeIdsParam)
	defer rows.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"userId": userId,
			"places": placeIdsParam,
		}).Error("Error getting likes for userId")
		return nil, err
	}
	for rows.Next() {
		var placeID uint
		var isLiked bool
		err := rows.Scan(&placeID, &isLiked)
		if err != nil {
			log.WithField("error", err).Error("Error reading like row")
			continue
		}
		result[placeID] = isLiked
	}
	return result, nil
}
