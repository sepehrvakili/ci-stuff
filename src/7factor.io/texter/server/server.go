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
	mongoCreds := texter.GetMongoCredsFromEnv()
	rrnDB, err := texter.NewMongoDB(mongoCreds)
	if err != nil {
		log.Fatalf("Unable to get a DB session: %v", err)
	}

	akCreds := texter.GetAKCredsFromEnv()
	ak, err := texter.NewAK(akCreds)
	if err != nil {
		log.Fatalf("Unable to get a DB session: %v", err)
	}

	defer ak.Close()
	defer rrnDB.Close()

	merger := texter.NewTagMerger()
	api := texter.NewAPI()
	cm := texter.NewCM(ak, rrnDB, api)

	// Please note that patterns for the URLs below must match
	// EXACTLY, including no trailing slashes.
	router := vestigo.NewRouter()
	router.Get("/status", texter.HealthCheck())
	router.Post("/messages/intro", texter.SendIntroSMS(twilio, rrnDB))
	router.Post("/campaigns", texter.StartCampaign(cm))
	router.Delete("/campaigns/:objId", texter.StopCampaign(cm))
	router.Post("/campaigns/:id/approve", texter.RequestApproval(twilio, rrnDB))
	router.Post("/campaigns/test", texter.TestCampaign(twilio, ak, merger))
	router.Post("/incoming", texter.Incoming(api, rrnDB, ak))
	log.Fatal(http.ListenAndServe(":"+port, router))
}
