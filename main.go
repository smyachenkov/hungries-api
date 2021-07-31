package main

import (
	"cloud.google.com/go/storage"
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"net/http"
	"os"
)

var db *sql.DB
var Dao *DaoEnv

type DaoEnv struct {
	PlacesDB     dao.PlaceDbService
	LikesDB      dao.LikeDBService
	MapsApi      dao.GoogleMapsAPIService
	CloudStorage dao.GoogleCloudStorageService
}

func initDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	return db.Ping()
}

func main() {
	// check required variables
	port := checkEnvVariable("PORT")
	databaseUrl := checkEnvVariable("DATABASE_URL")
	googleMapsApiKey := checkEnvVariable("GOOGLE_MAPS_API_KEY")
	storageKeyJson := checkEnvVariable("STORAGE_KEY_JSON")
	apiUsername := checkEnvVariable("API_USERNAME")
	apiPassword := checkEnvVariable("API_PASSWORD")

	// init DB and DAO objects
	err := initDB(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	mapsClient, _ := maps.NewClient(maps.WithAPIKey(googleMapsApiKey))
	cloudStorageClient, _ := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(storageKeyJson)))
	Dao = &DaoEnv{
		PlacesDB:     dao.PlaceDbService{DB: db},
		LikesDB:      dao.LikeDBService{DB: db},
		MapsApi:      dao.GoogleMapsAPIService{MapsClient: mapsClient},
		CloudStorage: dao.GoogleCloudStorageService{StorageClient: cloudStorageClient},
	}

	// run migrations
	m, err := migrate.New("file://db/migrations", databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil && err.Error() != "no change" {
		log.Fatal(err)
	}

	// set up routing
	router := mux.NewRouter()

	router.HandleFunc(
		"/places",
		BasicAuth(findNearbyPlacesHandler, apiUsername, apiPassword),
	).Methods(http.MethodGet)

	router.HandleFunc(
		"/places/liked",
		BasicAuth(getLikedPlacesHandler, apiUsername, apiPassword),
	).Methods(http.MethodGet)

	router.HandleFunc(
		"/place/{place}/like/{device}/{liked}",
		BasicAuth(saveLikeHandler, apiUsername, apiPassword),
	).Methods(http.MethodPost)

	http.ListenAndServe(":"+port, router)
}

func checkEnvVariable(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal("Missing $" + name + " environment variable")
	}
	return value
}
