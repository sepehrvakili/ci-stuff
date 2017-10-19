package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
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

func Test_SendMessageIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip()
	}

	port := os.Getenv("CONT_PORT")
	recipient := `{ "phonenumber" : "7067815146", "usdistrict" : "GA_07", "body" : "Your representative is {{targets.full_name}}" }`
	response, err := http.Post("http://localhost:"+port+"/messages/",
		"application/json", strings.NewReader(recipient))

	if err != nil {
		context.Fatalf("Received an error when POSTing message: %v", err.Error())
	}

	if status := response.StatusCode; status != http.StatusCreated {
		context.Errorf("Handler returned an incorrect status code: got %v want %v",
			status, http.StatusCreated)
	}
}
