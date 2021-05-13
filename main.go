package main

import (
	"database/sql"
	"encoding/json"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"googlemaps.github.io/maps"
	"hungries-api/dao"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var db *sql.DB
var Dao *DaoEnv

type DaoEnv struct {
	PlacesDB dao.PlaceDbService
	MapsApi  dao.GoogleMapsAPIService
}

func initDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	return db.Ping()
}

func findNearbyPlacesHandler(w http.ResponseWriter, r *http.Request) {
	pageTokenParam, hasToken := r.URL.Query()["pagetoken"]
	var pageToken string
	if hasToken {
		pageToken = pageTokenParam[0]
	}

	radius, _ := strconv.ParseUint(r.URL.Query()["radius"][0], 10, 64)

	coordinatesParam, hasCoordinates := r.URL.Query()["coordinates"]
	var coordinates maps.LatLng
	if hasCoordinates {
		latitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[0], 64)
		longitude, _ := strconv.ParseFloat(strings.Split(coordinatesParam[0], ",")[1], 64)
		coordinates = maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		}
	}

	places, err := FindNearbyPlaces(coordinates, uint(radius), pageToken)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Default().Print("Error discovering places " + err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func main() {
	// check required variables
	port := checkEnvVariable("PORT")
	databaseUrl := checkEnvVariable("DATABASE_URL")
	googleMapsApiKey := checkEnvVariable("GOOGLE_MAPS_API_KEY")

	// init DB and DAO objects
	err := initDB(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	mapsClient, _ := maps.NewClient(maps.WithAPIKey(googleMapsApiKey))
	Dao = &DaoEnv{
		PlacesDB: dao.PlaceDbService{DB: db},
		MapsApi:  dao.GoogleMapsAPIService{MapsClient: mapsClient},
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
	router.HandleFunc("/places", findNearbyPlacesHandler).Methods(http.MethodGet)
	http.ListenAndServe(":"+port, router)
}

func checkEnvVariable(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal("Missing $" + name + " environment variable")
	}
	return value
}
