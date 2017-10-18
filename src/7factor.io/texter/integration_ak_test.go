package texter

import (
	"strings"
	"testing"
)

func Test_GetCurrentSubscribersIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	targets, err := ak.GetCurrentSubscribers()
	if err != nil {
		context.Errorf("Something went wrong when querying AK: %v", err.Error())
	}

	// Assertions
	if len(targets) != 26 {
		context.Errorf("Expected %v subscribers but got %v", 26, len(targets))
	}

	for _, t := range targets {
		if len(t.PhoneNumber) != 10 || strings.HasPrefix(t.PhoneNumber, "1") {
			context.Errorf("Phone number incorrect format. got %v for user %v", t.PhoneNumber, t)
		}
	}
}

func Test_GetRepresentativesIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	reps, err := ak.GetRepresentatives()
	if err != nil {
		context.Errorf("Something went wrong when querying AK: %v", err.Error())
	}

	// Assertions
	if len(reps) != 435 {
		context.Errorf("Expected %v reps but got %v", 435, len(reps))
	}

	for _, r := range reps {
		if len(r.PhoneNumber) != 10 || strings.HasPrefix(r.PhoneNumber, "1") {
			context.Errorf("Phone number incorrect format. got %v for rep %v", r.PhoneNumber, r)
		}
	}
}

func Test_GuessDistrictForZipIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	district, err := ak.GuessDistrictForZip("22334")
	if err != nil {
		context.Errorf("Something went wrong when querying AK: %v", err.Error())
	}

	// Assertions
	if district != "VA_08" {
		context.Errorf("Expected %v reps but got %v", "VA_08", district)
	}
}

func Test_GuessDistrictForZipWithInvalidInputIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	district, err := ak.GuessDistrictForZip("")
	if err == nil {
		context.Errorf("Something should have went wrong when querying AK but it didn't. Got district %v", district)
	}
}

func Test_GetEmailsForPhoneNumberIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	emails, err := ak.GetEmailsForPhoneNumber("3423431411")
	if err != nil {
		context.Errorf("Something went wrong when querying AK: %v", err.Error())
	}

	// Assertions
	if len(emails) != 1 {
		context.Fatalf("Expected %v emails but got %v", 1, len(emails))
	}

	if email := emails[0]; email != "Aktest_907@stoptexasoil.com" {
		context.Errorf("Email is off. got %v want %v", email, "Aktest_907@stoptexasoil.com")
	}
}

func Test_GetEmailsForPhoneNumberWithInvalidInputIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetAKCredsFromEnv()
	ak, err := NewAK(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer ak.Close()

	emails, err := ak.GetEmailsForPhoneNumber("")
	if err == nil && len(emails) != 0 {
		context.Errorf("Something should have went wrong when querying AK. Got emails %v", emails)
	}
}
