package texter

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/husobee/vestigo"
)

// TestCampaign will send a test message to the target recipient
func TestCampaign(client TwilioClient, ak AK, m TagMerger) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		var test TestMessage
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&test)

		// Make sure we're going to close the request body when we're done.
		defer request.Body.Close()

		if decoderErr != nil {
			renderErrorWithHTTPStatus(writer, decoderErr, http.StatusBadRequest)
			return
		}

		// Guess district, err is returned if something goes horribly wrong with the query
		var mergedBody string
		district, err := ak.GuessDistrictForZip(test.ZipCode)
		if err != nil {
			// Replace with dummy text
			mergedBody = m.MergeUnknown(test.Body)
		} else {
			// Get reps, err will return if something is horribly wrong with the DB
			reps, err := ak.GetRepresentatives()
			if err != nil {
				renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
				return
			}

			// Grab the rep
			if repInfo, ok := reps[district]; ok {
				mergedBody = m.MergeRep(test.Body, repInfo)
			} else {
				mergedBody = m.MergeUnknown(test.Body)
			}
		}

		// The from number always lives in the environment
		from := CleanPhoneNumber(os.Getenv("TW_FROM"))

		// Simple translation of the recipient to a message.
		msg := Message{
			To:   CleanPhoneNumber(test.PhoneNumber),
			From: from,
			Body: mergedBody}

		status, smsErr := client.SendSMS(&msg)
		if smsErr != nil {
			renderError(writer, smsErr)
			return
		}

		writer.WriteHeader(status)
	}
}

// StartCampaign starts a campaign
func StartCampaign(cm CampaignManager) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		var campaign Campaign
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&campaign)

		// Make sure we're going to close the request body when we're done.
		defer request.Body.Close()

		if decoderErr != nil {
			renderErrorWithHTTPStatus(writer, decoderErr, http.StatusBadRequest)
			return
		}

		err := cm.Start(campaign)
		if err != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}

// StopCampaign stops a campaign
func StopCampaign(cm CampaignManager) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		campaignID := vestigo.Param(request, "objId")
		cm.Stop(campaignID)
		writer.WriteHeader(http.StatusOK)
	}
}
