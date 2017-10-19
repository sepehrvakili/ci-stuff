package main

import (
	"log"
	"net/http"
	"os"

	"7factor.io/texter"
	"github.com/husobee/vestigo"
)

// Listen and go.
func main() {
	port := os.Getenv("CONT_PORT")
	log.Printf("We're alive and well on port %s", port)

	twilio := texter.NewTW()
	merger := texter.NewTagMerger()

	mongoCreds := texter.GetMongoCredsFromEnv()
	messageStore, err := texter.NewMongoDB(mongoCreds)
	if err != nil {
		log.Fatalf("Unable to get a Mongo DB session: %v", err)
	}

	congressCreds := texter.GetCongressDBCredsFromEnv()
	congressdb, err := texter.NewCongressDB(congressCreds)
	if err != nil {
		log.Fatalf("Unable to get a Congress DB session: %v", err)
	}

	defer congressdb.Close()
	defer messageStore.Close()

	// Please note that patterns for the URLs below must match
	// EXACTLY, including no trailing slashes.
	router := vestigo.NewRouter()
	router.Get("/status", texter.HealthCheck())
	router.Post("/messages/", texter.SendMessage(merger, twilio, congressdb, messageStore))
	log.Fatal(http.ListenAndServe(":"+port, router))
}
