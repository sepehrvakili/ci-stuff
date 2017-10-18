package texter

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// Incoming handles incoming SMS messages from Twilio
func Incoming(api RRNAPI, db DB, ak AK) http.HandlerFunc {
	// Type coercion black magic.
	return func(writer http.ResponseWriter, request *http.Request) {
		// forever write OK, Twilio is a honey badger
		writer.WriteHeader(http.StatusCreated)
		writer.Header().Set("Content-Type", "application/json")

		if request.Body == nil {
			log.Printf("request has no body, we can do nothing with this")
			return
		}

		defer request.Body.Close()

		// Read bytes
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("something went horribly wrong when reading body")
			return
		}

		// Twilio's crap is form encoded
		values, err := url.ParseQuery(string(bytes))
		if err != nil {
			log.Printf("unable to decode the twilio message body %+v", string(bytes))
		}

		message := NewTwilioMessage(values)
		process(message, api, db, ak)
	}
}

// Regex patterns for detecting  messages.
var approvalMatcher = regexp.MustCompile(`^[\s]*(?P<id>[\d]+)[\s]*$`)
var startMatcher = regexp.MustCompile(`(?i)^\s*(START|YES|UNSTOP)\s*$`)
var stopMatcher = regexp.MustCompile(`(?i)^\s*(STOP|STOPALL|UNSUBSCRIBE|CANCEL|END|QUIT)\s*$`)

func process(message TwilioMessage, api RRNAPI, db DB, ak AK) {
	log.Printf("Processing message %+v", message)
	var err error
	if m := approvalMatcher.FindStringSubmatch(message.Body); m != nil && len(m) > 0 {
		cid, err := strconv.Atoi(m[1])
		if err != nil {
			log.Printf("something has gone horribly wrong, cannot parse as int %v", m[1])
		} else {
			err = isApprovalMessage(cid, message, api, db)
		}
	} else if m = startMatcher.FindStringSubmatch(message.Body); m != nil && len(m) > 0 {
		err = isStartMessage(message, api, ak)
	} else if m = stopMatcher.FindStringSubmatch(message.Body); m != nil && len(m) > 0 {
		err = isStopMessage(message, api, ak)
	} else {
		log.Printf("Not sure what to do with message %+v", message)
		err = stowMessage(message, db)
	}

	if err != nil {
		log.Printf("error processing twilio message %+v err: %v", message, err)
	}
}

func isApprovalMessage(cid int, message TwilioMessage, api RRNAPI, db DB) error {
	approval, err := db.GetMostRecentApprovalByID(cid)
	if err != nil {
		// Cannot continue
		return err
	}

	approver, err := findApprover(api, message.From, approval)
	if err != nil {
		// Cannot continue
		return err
	}

	err = api.ApproveCampaign(approval, approver)
	if err != nil {
		// Cannot continue
		return err
	}

	err = db.ApproveCampaign(approval, approver)
	if err != nil {
		// We were unable to record an approval, but the campaign is actually approved.
		// We're probably out of sync with the API, but that's OK. API is the source of
		// truth. Return true so we don't stow the message, but log the error
		log.Printf("unable to approve campaign %v because %v", approval, err)
	}

	return nil
}

func findApprover(api RRNAPI, phoneNumber string, approval ApprovalRequest) (Approver, error) {
	// try to get approvers from the API first, it's most up to date
	approvers, err := api.GetAllApprovers()
	if err != nil || len(approvers) == 0 {
		log.Printf("could not retrieve approver list from API, list size is %v and error is %v. using approvers on the original request", len(approvers), err)
		approvers = approval.Approvers
	}

	for _, approver := range approvers {
		if CleanPhoneNumber(approver.PhoneNumber) == CleanPhoneNumber(phoneNumber) {
			return approver, nil
		}
	}

	return Approver{}, fmt.Errorf("Requested Approver phone number not found in the list: %v",
		CleanPhoneNumber(phoneNumber))
}

func isStartMessage(message TwilioMessage, api RRNAPI, ak AK) error {
	emails, err := ak.GetEmailsForPhoneNumber(message.From)
	if err != nil {
		return err
	}

	err = api.Subscribe(emails)
	if err != nil {
		return err
	}

	return nil
}

func isStopMessage(message TwilioMessage, api RRNAPI, ak AK) error {
	emails, err := ak.GetEmailsForPhoneNumber(message.From)
	if err != nil {
		return err
	}

	err = api.Unsubscribe(emails)
	if err != nil {
		return err
	}

	return nil
}

func stowMessage(message TwilioMessage, db DB) error {
	err := db.StowMessage(message)
	if err != nil {
		return err
	}

	return nil
}
