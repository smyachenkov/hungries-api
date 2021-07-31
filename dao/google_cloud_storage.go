package dao

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type GoogleCloudStorageService struct {
	StorageClient *storage.Client
}

const placePhotosBucketName = "hungries-place-photo"

// UploadPhoto upload photo to bucket
func (s *GoogleCloudStorageService) UploadPhoto(placeId string, image io.ReadCloser) (string, error) {
	log.WithField("placeId", placeId).Info("Saving new photo for place")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// check if there is an object with that name
	existingObjects := s.StorageClient.Bucket(placePhotosBucketName).Objects(ctx, &storage.Query{Prefix: placeId})
	nextObject, _ := existingObjects.Next()
	if nextObject != nil {
		return getPublicUrl(placeId), nil
	}
	// upload object
	wc := s.StorageClient.Bucket(placePhotosBucketName).Object(placeId).NewWriter(ctx)
	if _, err := io.Copy(wc, image); err != nil {
		log.WithField("placeId", placeId).Info("Error uploading new photo")
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.WithField("placeId", placeId).Info("Error uploading new photo")
		return "", fmt.Errorf("Writer.Close: %v", err)
	}
	photoUrl := getPublicUrl(placeId)
	return photoUrl, nil
}

func getPublicUrl(placeId string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", placePhotosBucketName, placeId)
}
