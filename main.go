package main

import (
	"context"
	"encoding/json"
	"fmt"
	"geojosn-api/mongodb"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	fmt.Println("GeoJSON Server...")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	fmt.Println("Initialize chi")
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.Limit(3, time.Minute, httprate.WithKeyFuncs(
		httprate.KeyByIP,
		httprate.KeyByEndpoint,
	)))

	r.Use(middleware.Timeout(60 * time.Second))
	fmt.Println("Initialize chi finish...")

	fmt.Println("Initialize mongodb connection")
	err := mongodb.InitMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Initialize mongodb connection finish...")

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Selamat datang di server..."))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is working..."))
	})

	r.Get("/coffee_shops", getAllCofeeShop)
	// r.Post("/coffee_shops", getAllCofeeShop)
	// r.Post("/coffee_shops/near", getAllCofeeShop)

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "9000"
	}
	fmt.Printf("GeoJSON Server runnincleag on http://0.0.0.0:%s\n\n", appPort)

	// http.ListenAndServe("0.0.0.0:"+appPort, r)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+appPort, r))
	// log.Fatal(http.ListenAndServe(":"+appPort, r))

}

func getAllCofeeShop(w http.ResponseWriter, r *http.Request) {

	coll := mongodb.Client.Database(os.Getenv("DB_NAME")).Collection("coffee_shop")

	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	if len(results) == 0 {
		fmt.Printf("No documents were found\n")
		return
	}

	geoJSONResponse := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": results,
	}

	responseData, err := json.MarshalIndent(geoJSONResponse, "", "    ")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

// func getCoffeeShopByName(w http.ResponseWriter, r *http.Request) {

// 	coll := mongodb.Client.Database(os.Getenv("DB_NAME")).Collection("coffee_shop")

// 	cursor, err := coll.Find(context.TODO(), bson.D{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer cursor.Close(context.TODO())

// 	var results []bson.M
// 	if err = cursor.All(context.TODO(), &results); err != nil {
// 		panic(err)
// 	}

// 	if len(results) == 0 {
// 		fmt.Printf("No documents were found\n")
// 		return
// 	}

// 	jsonData, err := json.MarshalIndent(results, "", "    ")
// 	if err != nil {
// 		panic(err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(jsonData)
// }
