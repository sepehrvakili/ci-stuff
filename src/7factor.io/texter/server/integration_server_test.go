package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func Test_StatusEndpointIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	port := os.Getenv("CONT_PORT")
	response, err := http.Get("http://localhost:" + port + "/status")

	if err != nil {
		context.Fatalf("Received an error when pulling the health check: %v", err.Error())
	}

	if status := response.StatusCode; status != http.StatusOK {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"gtg":true}`
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		context.Errorf("Received an error when reading response body: %v", err.Error())
	}

	if string(body) != expected {
		context.Errorf("Handler returned unexpected body: got %v want %v",
			body, expected)
	}
}

func Test_SendIntroSMSIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	recipient := `{ "name": "jduv", "PhoneNumber":"7067815146", "Email":"jduvall@credoaction.com" }`
	response, err := http.Post("http://localhost:"+port+"/messages/intro",
		"application/json", strings.NewReader(recipient))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	if status := response.StatusCode; status != http.StatusCreated {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func Test_RequestApprovalIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	campaignID := "1"
	objectID := bson.NewObjectId()
	approvalTemplate := `{ "_id": "%v", "body":"Test approval message","version":1, "approvers":[ { "name":"jduv", "phoneNumber":"+17067815146","email":"jduvall@credoaction.com" }, {"name":"jduv", "phoneNumber":"17067815146", "email":"jduvall@credoaction.com" }]}`

	response, err := http.Post("http://localhost:"+port+"/campaigns/"+campaignID+"/approve",
		"application/json", strings.NewReader(fmt.Sprintf(approvalTemplate, objectID.Hex())))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	if status := response.StatusCode; status != http.StatusAccepted {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusAccepted)
	}
}

func Test_CampaignTestIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	test := `{ "phoneNumber": "7067815146", "zipCode" : "22334", "body": 
		"Call {{ targets.title }} {{ targets.full_name }} at {{ targets.phone }}" }`
	response, err := http.Post("http://localhost:"+port+"/campaigns/test",
		"application/json", strings.NewReader(test))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	if status := response.StatusCode; status != http.StatusCreated {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func Test_CampaignTestWithBadZipIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	test := `{ "phoneNumber": "7067815146", "zipCode" : "ABCDEF", "body": 
		"Call {{ targets.title }} {{ targets.full_name }} at {{ targets.phone }}" }`
	response, err := http.Post("http://localhost:"+port+"/campaigns/test",
		"application/json", strings.NewReader(test))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	// This should still work, we just will replace tags with default
	// values. That will be tested in a unit test, we just want to ensure
	// that the server doesn't break if you send it bad information.
	if status := response.StatusCode; status != http.StatusCreated {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func Test_IncomingAlwaysReturnsCreatedIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	test := `{  "MessageSid":  "ABC123",
				"AccountSid": "ABC123",
				"MessagingServiceSid": "ABC123",
				"From": "4045551234",
				"To": "4045551111",
				"Body": "My phone number is 555-404-1234",
				"NumMedia": 0 }`

	response, err := http.Post("http://localhost:"+port+"/incoming",
		"application/json", strings.NewReader(test))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	// This should still work, we just will replace tags with default
	// values. That will be tested in a unit test, we just want to ensure
	// that the server doesn't break if you send it bad information.
	if status := response.StatusCode; status != http.StatusCreated {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusCreated)
	}
}
