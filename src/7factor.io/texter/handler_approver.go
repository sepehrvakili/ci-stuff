package texter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/husobee/vestigo"
)

// RequestApproval Handler for sending a welcome SMS to the person in the
// payload. Will utilize twilio.
func RequestApproval(client TwilioClient, db DB) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		var approval ApprovalRequest
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&approval)

		// Make sure we're going to close the request body when we're done.
		defer request.Body.Close()

		if decoderErr != nil {
			renderErrorWithHTTPStatus(writer, decoderErr, http.StatusBadRequest)
			return
		}

		// The from number always lives in the environment
		from := CleanPhoneNumber(os.Getenv("TW_FROM"))
		campaignID := vestigo.Param(request, "id")

		// Grab the approval template from the DB
		template, dbErr := db.GetMessageTemplateByKey("approval-sms")
		if dbErr != nil {
			renderError(writer, dbErr)
			return
		}

		collector := ErrorCollector{}
		body := fmt.Sprintf(template.Body, approval.Body, campaignID)
		for i, approver := range approval.Approvers {
			// While we're rolling, clean phone numbers since they'll be going
			// into our database. All external #s should be scrubbed
			approver.PhoneNumber = CleanPhoneNumber(approver.PhoneNumber)
			approval.Approvers[i] = approver

			msg := Message{
				To:   CleanPhoneNumber(approver.PhoneNumber),
				From: from,
				Body: body}

			_, smsErr := client.SendSMS(&msg)
			if smsErr != nil {
				collector.Collect(smsErr)
			}
		}

		if len(collector) == len(approval.Approvers) {
			log.Printf("All messages we attempted to send were errors")
			renderErrorWithHTTPStatus(writer, &collector, http.StatusBadRequest)
			return
		}

		// Set the CID from the path params
		id, err := strconv.Atoi(campaignID)
		if err != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusBadRequest)
			return
		}
		approval.CampaignID = id

		err = db.RegisterApproval(approval)
		if err != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
	}
}
