package texter

import (
	"fmt"
	"log"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func clearCollection(collection string) {
	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		panic(fmt.Errorf("Cannot clean up after ourselves. Cannot get a DB %v", err.Error()))
	}

	log.Printf("Cleaning up after ourselves by nuking collection %v", collection)
	db.(MongoDB).Session.DB(creds.DBName).C(collection).RemoveAll(nil)
}

func Test_StowMessageIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	expected := Message{
		From: "4045551234",
		To:   "4045551234",
		Body: "Something to Stow",
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("Messages")

	err = db.StowMessage(expected)
	if err != nil {
		context.Fatalf("Unable to stow message. Error %v", err)
	}

	var actual Message
	err = db.(MongoDB).Session.DB(creds.DBName).C("Messages").Find(bson.M{
		"body": expected.Body,
	}).One(&actual)

	if err != nil {
		context.Fatalf("Unable to insert test record %v", err)
	}

	if actual.From != expected.From {
		context.Fatalf("Field doesn't match. got %v want %v", actual.From, expected.From)
	}

	if actual.To != expected.To {
		context.Fatalf("Field doesn't match. got %v want %v", actual.To, expected.To)
	}

	if actual.Body != expected.Body {
		context.Fatalf("Field doesn't match. got %v want %v", actual.Body, expected.Body)
	}
}
