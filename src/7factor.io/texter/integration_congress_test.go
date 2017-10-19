package texter

import (
	"strings"
	"testing"
)

func Test_GetRepresentativesIntegration(context *testing.T) {
	if testing.Short() {
		context.Skip("Skipping integration tests")
	}

	creds := GetCongressDBCredsFromEnv()
	db, err := NewCongressDB(creds)
	if err != nil {
		context.Errorf("Something went wrong when getting a connection: %v", err.Error())
	}

	defer db.Close()

	reps, err := db.GetRepresentatives()
	if err != nil {
		context.Errorf("Something went wrong when querying CongressDB: %v", err.Error())
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
