package texter

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var personOfInterest = `
{
	"name": "Jduv",
	"phoneNumber": "7067815146",
	"email": "jduv@7factor.io"
}
`

var badPhoneNumber = `
{
	"name": "Jduv",
	"phoneNumber": "ABCDEFJKE24",
	"email": "jduv@7factor.io"
}
`

func Test_SendIntroSMSWithValidTarget(context *testing.T) {
	// The From field is gleaned from the environment so
	// let's set it so we can inspect
	expectedFrom := "4045551234"
	os.Setenv("TW_FROM", expectedFrom)

	var mockClient = new(MockTwilioClient)
	var capturedMsg *Message
	mockClient.MockSendSMSMethod = func(msg *Message) (int, error) {
		capturedMsg = msg
		return http.StatusCreated, nil
	}

	var mockDB = new(MockDB)
	mockDB.MockGetMessageTemplateByKey = func(key string) (MessageTemplate, error) {
		return MessageTemplate{
			ID:     "ABCD",
			Key:    key,
			Body:   "Hello World",
			Active: true}, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/messages/intro-sms",
		"/messages/intro-sms",
		bytes.NewBufferString(personOfInterest),
		SendIntroSMS(mockClient, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if mockClient.SendSMSCallCount != 1 {
				context.Errorf("Unexpected interaction: Call cound should be %v got %v",
					1, mockClient.SendSMSCallCount)
			}
			if status := recorder.Code; status != http.StatusCreated {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusCreated)
			}
			if capturedMsg.From != expectedFrom {
				context.Errorf("Captured message had an incorrect from number: got %v want %v",
					capturedMsg.From, expectedFrom)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendIntroSMSWithBadJSON(context *testing.T) {
	var mockDB = new(MockDB)
	var mockClient = new(MockTwilioClient)
	theTest := TestParameters{
		http.MethodPost,
		"/messages/intro-sms",
		"/messages/intro-sms",
		bytes.NewBufferString(`{ {{ : d34 ;lks }}}`),
		SendIntroSMS(mockClient, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			// Assert
			if status := recorder.Code; status != http.StatusBadRequest {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusBadRequest)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendIntroSMSHandlesTwilioErrors(context *testing.T) {
	var mockClient = new(MockTwilioClient)
	var mockDB = new(MockDB)
	mockDB.MockGetMessageTemplateByKey = func(key string) (MessageTemplate, error) {
		return MessageTemplate{
			ID:     "ABCD",
			Key:    key,
			Body:   "Hello World",
			Active: true}, nil
	}
	expectedStatus := 403
	exception := TwilioException{
		Status:   expectedStatus,
		Message:  "Bad juju",
		Code:     9999,
		MoreInfo: "Badder juju",
	}

	mockClient.MockSendSMSMethod = func(message *Message) (int, error) {
		return expectedStatus, exception
	}
	theTest := TestParameters{
		http.MethodPost,
		"/messages/intro-sms",
		"/messages/intro-sms",
		bytes.NewBufferString(personOfInterest),
		SendIntroSMS(mockClient, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			// Assert
			if mockClient.SendSMSCallCount != 1 {
				context.Errorf("Unexpected interaction: Call cound should be %v got %v",
					1, mockClient.SendSMSCallCount)
			}
			if status := recorder.Code; status != expectedStatus && status != exception.Status {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, exception.Status)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendIntroSMSHandlesDBError(context *testing.T) {
	var mockDB = new(MockDB)
	mockDB.MockGetMessageTemplateByKey = func(key string) (MessageTemplate, error) {
		return MessageTemplate{}, errors.New("Unable to access the database")
	}

	var mockClient = new(MockTwilioClient)
	theTest := TestParameters{
		http.MethodPost,
		"/messages/intro-sms",
		"/messages/intro-sms",
		bytes.NewBufferString(personOfInterest),
		SendIntroSMS(mockClient, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			// Assert
			if status := recorder.Code; status != http.StatusInternalServerError {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusInternalServerError)
			}
		},
	}
	Run(context, theTest)
}
