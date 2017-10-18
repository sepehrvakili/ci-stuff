package texter

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

var approval = `
{
	"body": "This is my test approval message",
	"version": 1,
	"approvers": [
		{
			"name":"jduv",
			"phoneNumber":"+17067815146",
			"email":"jduv@7factor.io"
		},{
			"name":"jduv",
			"phoneNumber":"+17067815146",
			"email":"jduv@7factor.io"
		}
	]
}
`

func Test_RequestApprovalWithValidRequest(context *testing.T) {
	mockDB := new(MockDB)
	var capturedApproval ApprovalRequest
	mockDB.MockRegisterApproval = func(approval ApprovalRequest) error {
		capturedApproval = approval
		return nil
	}
	mockDB.MockGetMessageTemplateByKey = func(key string) (MessageTemplate, error) {
		return MessageTemplate{
			ID:     "ABCDEF",
			Key:    key,
			Body:   "Hello World",
			Active: true}, nil
	}

	mockTwilio := new(MockTwilioClient)
	var capturedMessage *Message
	mockTwilio.MockSendSMSMethod = func(msg *Message) (int, error) {
		capturedMessage = msg
		return 200, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/campaigns/123/approve",
		"/campaigns/:id/approve",
		bytes.NewBufferString(approval),
		RequestApproval(mockTwilio, mockDB),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if code := recorder.Code; code != http.StatusAccepted {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					code, http.StatusAccepted)
			}
			if mockTwilio.SendSMSCallCount != 2 {
				context.Errorf("Handler did not send enough SMS messages: got %v want %v",
					mockTwilio.SendSMSCallCount, 2)
			}
			if mockDB.RegisterApprovalCallCount != 1 {
				context.Errorf("Handler did not write to Mongo as expected: got %v want %v",
					mockTwilio.SendSMSCallCount, 1)
			}

			// Test to see if we clean #s before sending to the DB
			for _, a := range capturedApproval.Approvers {
				if a.PhoneNumber != "7067815146" {
					context.Errorf("Did not get the phone number we expected. got %v want %v",
						a.PhoneNumber, "7067815146")
				}
			}
		},
	}
	Run(context, theTest)
}

func Test_RequestApprovalWithBadJSON(context *testing.T) {
	var mockDB = new(MockDB)
	var mockTwilio = new(MockTwilioClient)
	theTest := TestParameters{
		http.MethodPost,
		"/campaigns/123/approve",
		"/campaigns/:id/approve",
		bytes.NewBufferString(`{ {{ : d34 ;lks }}}`),
		RequestApproval(mockTwilio, mockDB),
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
