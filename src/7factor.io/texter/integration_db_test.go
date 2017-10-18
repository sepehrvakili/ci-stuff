package texter

import (
	"fmt"
	"log"
	"testing"
	"time"

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

func Test_GetAllMessageTemplatesIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	list, err := db.GetActiveMessageTemplates()
	if err != nil {
		context.Errorf("Unable to get a list of message templates: %v", err.Error())
	}

	if len(list) <= 0 {
		context.Errorf("No items in the list!")
	}

	for _, element := range list {
		if !element.Active {
			context.Errorf("Got an inactive element: %v", element)
		}
	}
}

func Test_GetMessageTemplateByKeyIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	template, err := db.GetMessageTemplateByKey("intro-sms")
	if err != nil {
		context.Fatalf("Unable to get the requested message template: %v", err.Error())
	}

	if template.Key != "intro-sms" {
		context.Errorf("Got the wrong message template: %v", template)
	}
}

func Test_RegisterApproverIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	campaignID := 1
	expected := ApprovalRequest{
		ID:          bson.NewObjectId(),
		Body:        "Test Body",
		CampaignID:  campaignID,
		RequestedAt: time.Now(),
		Version:     1,
		Approvers: []Approver{
			{
				Name:        "jduv",
				PhoneNumber: "7065551234",
				Email:       "jduv@7factor.io",
			},
		},
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	err = db.RegisterApproval(expected)
	if err != nil {
		context.Fatalf("Unable to register approval: %v", err.Error())
	}

	var actual ApprovalRequest
	query := bson.M{
		"campaignID": campaignID,
	}
	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(query).One(&actual)
	if err != nil {
		context.Fatalf("Problem when reading from ApprovalRequests collection: %v", err.Error())
	}
}

func Test_RegisterApproverUpsertDuplicateIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	campaignID := 1
	expected := ApprovalRequest{
		ID:         bson.NewObjectId(),
		CampaignID: campaignID,
		Body:       "Test Body",
		Version:    1,
		Approvers: []Approver{
			{
				Name:        "jduv",
				PhoneNumber: "7065551234",
				Email:       "jduv@7factor.io",
			},
		},
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	// Insert twice the exact same record.
	duplicates := 1
	for i := 0; i < duplicates; i++ {
		err = db.RegisterApproval(expected)
		if err != nil {
			context.Fatalf("Unable to register approval: %v", err.Error())
		}
	}

	query := bson.M{
		"campaignID": campaignID,
		"version":    1,
	}
	count, err := db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(query).Count()
	if err != nil {
		context.Fatalf("Problem when reading from ApprovalRequests collection: %v", err.Error())
	}
	if count > 1 {
		context.Fatalf("Duplicate approval requests in the DB. Found %v instances", count)
	}
}

func Test_RegisterApprovalUpsertIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	campaignID := 1
	expected := ApprovalRequest{
		ID:         bson.NewObjectId(),
		CampaignID: campaignID,
		Body:       "Test Body",
		Version:    1,
		Approvers: []Approver{
			{
				Name:        "jduv",
				PhoneNumber: "7065551234",
				Email:       "jduv@7factor.io",
			},
		},
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	// Insert twice.
	numberOfInserts := 2
	for i := 1; i <= numberOfInserts; i++ {
		expected.Version = i
		err = db.RegisterApproval(expected)
		if err != nil {
			context.Fatalf("Unable to register approval: %v", err.Error())
		}
	}

	count, err := db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(bson.M{
		"_id": expected.ID,
	}).Count()
	if err != nil {
		context.Fatalf("Problem when counting ApprovalRequests collection: %v", err.Error())
	}

	if count != 1 {
		context.Fatalf("Expected exactly %v values for campaign with id %v. Found %v.",
			numberOfInserts, campaignID, count)
	}

	var actual ApprovalRequest
	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(bson.M{"campaignID": expected.CampaignID}).One(&actual)
	if err != nil {
		context.Fatalf("Problem when reading from ApprovalRequests collection: %v", err.Error())
	}

	if actual.Version != numberOfInserts {
		context.Fatalf("Expected version %v but got %v.", numberOfInserts, actual.Version)
	}

}

func Test_GetMostRecentApprovalByIDIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	expected := ApprovalRequest{
		ID:         bson.NewObjectId(),
		CampaignID: -1,
		Body:       "Test Body",
		Version:    1,
		Approvers: []Approver{
			{
				Name:        "jduv",
				PhoneNumber: "4045551234",
				Email:       "jduv@7factor.io",
			},
		},
	}

	// Insert a couple of records
	maxVersion := 3
	for i := 1; i <= maxVersion; i++ {
		expected.Version = i
		_, err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Upsert(bson.M{"_id": expected.ID}, expected)
		if err != nil {
			context.Fatalf("Unable to insert test record %v", err)
		}
	}

	actual, err := db.GetMostRecentApprovalByID(-1)
	if err != nil {
		context.Fatalf("Unable to retrieve test record %v", err)
	}

	if actual.Version != maxVersion {
		context.Fatalf("Versions do not match. got %v want %v", actual.Version, expected.Version)
	}

	if actual.CampaignID != expected.CampaignID {
		context.Fatalf("CampaignIDs do not match. got %v want %v", actual.CampaignID, expected.CampaignID)
	}

	if actual.ApprovedOn != nil {
		context.Fatalf("ApprovedOn should be nil, got %v", actual.ApprovedOn)
	}
}

func Test_ApproveCampaignWithExistingApproverIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	approver := Approver{
		Name:        "jduv",
		PhoneNumber: "4045551234",
		Email:       "jduv@7factor.io",
	}

	expected := ApprovalRequest{
		ID:          bson.NewObjectId(),
		CampaignID:  -1,
		RequestedAt: time.Now(),
		Body:        "Test Body",
		Version:     1,
		Approvers:   []Approver{approver},
	}

	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Insert(expected)
	if err != nil {
		context.Fatalf("Unable to insert test record %v", err)
	}

	err = db.ApproveCampaign(expected, approver)
	if err != nil {
		context.Fatalf("Unable to approve campaign %v with %v", expected, approver)
	}

	var actual ApprovalRequest
	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(
		bson.M{
			"campaignID": expected.CampaignID,
			"version":    expected.Version,
		}).One(&actual)

	if err != nil {
		context.Fatalf("Unable to retrieve the updated record")
	}

	if actual.Approvers[0].HasApproved != true {
		context.Errorf("Approval failed %v", actual)
	}

	if actual.ApprovedOn == nil {
		context.Errorf("ApprovedOn date should not be nil in %v", actual)
	}
}

func Test_ApproveCampaignWithAddedApproverIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("ApprovalRequests")

	newApprover := Approver{
		Name:        "jduv",
		PhoneNumber: "4045551234",
		Email:       "jduv@7factor.io",
	}

	expected := ApprovalRequest{
		ID:          bson.NewObjectId(),
		CampaignID:  -1,
		RequestedAt: time.Now(),
		Body:        "Test Body",
		Version:     1,
		Approvers:   []Approver{},
	}

	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Insert(expected)
	if err != nil {
		context.Fatalf("Unable to insert test record %v", err)
	}

	err = db.ApproveCampaign(expected, newApprover)
	if err != nil {
		context.Fatalf("Unable to approve campaign %v with %v", expected, newApprover)
	}

	var actual ApprovalRequest
	err = db.(MongoDB).Session.DB(creds.DBName).C("ApprovalRequests").Find(
		bson.M{
			"campaignID": expected.CampaignID,
			"version":    expected.Version,
		}).One(&actual)

	if err != nil {
		context.Fatalf("Unable to retrieve the updated record")
	}

	if len(actual.Approvers) != 1 {
		context.Errorf("Should be one new approver, got %v", len(actual.Approvers))
	}

	if actual.Approvers[0].HasApproved != true {
		context.Errorf("Approval failed %v", actual)
	}

	if actual.ApprovedOn == nil {
		context.Errorf("ApprovedOn date should not be nil in %v", actual)
	}
}

func Test_StowMessageIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	expected := TwilioMessage{
		MessageSid: "ABC123",
		AccountSid: "ABC123",
		From:       "4045551234",
		To:         "4045551234",
		Body:       "Something to Stow",
	}

	creds := GetMongoCredsFromEnv()
	db, err := NewMongoDB(creds)
	if err != nil {
		context.Fatalf("Unable to get a database instance: %v", err.Error())
	}

	defer db.Close()
	defer clearCollection("LostAndFound")

	err = db.StowMessage(expected)
	if err != nil {
		context.Fatalf("Unable to stow message. Error %v", err)
	}

	var actual TwilioMessage
	err = db.(MongoDB).Session.DB(creds.DBName).C("LostAndFound").Find(bson.M{
		"body": expected.Body,
	}).One(&actual)
	if err != nil {
		context.Fatalf("Unable to insert test record %v", err)
	}

	if actual.AccountSid != expected.AccountSid {
		context.Fatalf("Field doesn't match. got %v want %v", actual.AccountSid, expected.AccountSid)
	}

	if actual.MessageSid != expected.MessageSid {
		context.Fatalf("Field doesn't match. got %v want %v", actual.MessageSid, expected.MessageSid)
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
