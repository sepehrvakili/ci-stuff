package texter

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var target = `
{
	"phoneNumber": "7067815146",
	"usdistrict" : "GA_7",
	"body" : "This is a test"
}
`

func Test_SendMessageWithValidInformationButNoReplacements(context *testing.T) {
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
	mockDB.MockStowMessage = func(message Message) error {
		return nil
	}

	// Just fake this, we don't care about the assertion for now.
	var mockCongressDB = new(MockCongressDB)
	mockCongressDB.MockGetRepresentatives = func() (Representatives, error) {
		return map[string]RepInfo{
			"GA_7": {},
		}, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/messages/",
		"/messages/",
		bytes.NewBufferString(target),
		SendMessage(NewTagMerger(), mockClient, mockCongressDB, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
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

func Test_SendMessageWithDBErrorFails(context *testing.T) {
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
	mockDB.MockStowMessage = func(message Message) error {
		return errors.New("we esplody")
	}

	// Just fake this, we don't care about the assertion for now.
	var mockCongressDB = new(MockCongressDB)
	mockCongressDB.MockGetRepresentatives = func() (Representatives, error) {
		return map[string]RepInfo{
			"GA_7": {},
		}, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/messages/",
		"/messages/",
		bytes.NewBufferString(target),
		SendMessage(NewTagMerger(), mockClient, mockCongressDB, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusInternalServerError {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusInternalServerError)
			}
			if capturedMsg.From != expectedFrom {
				context.Errorf("Captured message had an incorrect from number: got %v want %v",
					capturedMsg.From, expectedFrom)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendMessageWithBadJSONFails(context *testing.T) {
	// The From field is gleaned from the environment so
	// let's set it so we can inspect
	expectedFrom := "4045551234"
	os.Setenv("TW_FROM", expectedFrom)

	var mockClient = new(MockTwilioClient)
	var mockDB = new(MockDB)
	var mockCongressDB = new(MockCongressDB)

	theTest := TestParameters{
		http.MethodPost,
		"/messages/",
		"/messages/",
		bytes.NewBufferString("{asdllke;lkj:kj}"),
		SendMessage(NewTagMerger(), mockClient, mockCongressDB, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusBadRequest {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusCreated)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendMessageWithCongressDBErrorFails(context *testing.T) {
	// The From field is gleaned from the environment so
	// let's set it so we can inspect
	expectedFrom := "4045551234"
	os.Setenv("TW_FROM", expectedFrom)

	var mockClient = new(MockTwilioClient)
	var mockDB = new(MockDB)
	var mockCongressDB = new(MockCongressDB)
	mockCongressDB.MockGetRepresentatives = func() (Representatives, error) {
		return nil, errors.New("we esplody")
	}

	theTest := TestParameters{
		http.MethodPost,
		"/messages/",
		"/messages/",
		bytes.NewBufferString("{}"),
		SendMessage(NewTagMerger(), mockClient, mockCongressDB, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusInternalServerError {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusInternalServerError)
			}
		},
	}
	Run(context, theTest)
}

func Test_SendMessageHandlesTwilioErrors(context *testing.T) {
	var mockClient = new(MockTwilioClient)
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

	// Just fake this, we don't care about the assertion for now.
	var mockCongressDB = new(MockCongressDB)
	mockCongressDB.MockGetRepresentatives = func() (Representatives, error) {
		return map[string]RepInfo{
			"GA_7": {},
		}, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/messages/",
		"/messages/",
		bytes.NewBufferString(target),
		SendMessage(NewTagMerger(), mockClient, mockCongressDB, new(MockDB)),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != expectedStatus && status != exception.Status {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, exception.Status)
			}
		},
	}
	Run(context, theTest)
}
