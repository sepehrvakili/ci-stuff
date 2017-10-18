package texter

import (
	"encoding/json"
	"net/http"
	"os"
)

// SendIntroSMS Handler for sending a welcome SMS to the person in the
// payload. Will utilize twilio.
func SendIntroSMS(client TwilioClient, db DB) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		var to Recipient
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&to)

		// Make sure we're going to close the request body when we're done.
		defer request.Body.Close()

		if decoderErr != nil {
			renderErrorWithHTTPStatus(writer, decoderErr, http.StatusBadRequest)
			return
		}

		// The from number always lives in the environment
		from := os.Getenv("TW_FROM")

		template, dbErr := db.GetMessageTemplateByKey("intro-sms")
		if dbErr != nil {
			renderError(writer, dbErr)
			return
		}

		// Simple translation of the recipient to a message.
		msg := Message{
			To:   CleanPhoneNumber(to.PhoneNumber),
			From: CleanPhoneNumber(from),
			Body: template.Body}

		status, smsErr := client.SendSMS(&msg)
		if smsErr != nil {
			renderError(writer, smsErr)
			return
		}

		writer.WriteHeader(status)
	}
}
