package texter

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// SendMessage will send the payload via twilio and save the message to our messages
// history DB table.
func SendMessage(merger TagMerger, tw TwilioClient, congress Congress, db DB) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		var target Target
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&target)

		defer request.Body.Close()

		if err != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusBadRequest)
			return
		}

		// Pull twilio info from the environment
		from := os.Getenv("TW_FROM")

		// Grab reps
		reps, err := congress.GetRepresentatives()
		if err != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
			return
		}

		var mergedBody string
		rep, ok := reps[target.USDistrict]
		if !ok {
			log.Printf("Unable to look up representative for district %v", target.USDistrict)
			mergedBody = merger.MergeUnknown(target.Body)
		} else {
			mergedBody = merger.MergeRep(target.Body, rep)
		}

		// Simple translation of the recipient to a message.
		msg := Message{
			To:   CleanPhoneNumber(target.PhoneNumber),
			From: CleanPhoneNumber(from),
			Body: mergedBody,
		}

		status, err := tw.SendSMS(&msg)
		if err != nil {
			renderError(writer, err)
			return
		}

		err = db.StowMessage(msg)
		if err != nil {
			log.Printf("Unable to save message, but we did send it!")
			renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(status)
	}
}
