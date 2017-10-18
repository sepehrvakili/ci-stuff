package texter

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testCampaign = `
{
	"phoneNumber": "17067815146",
	"zipCode" : "30328",
	"body": "Hello from my test campaign"
}
`

func Test_SendTestMessageWithValidRequest(context *testing.T) {
	// The From field is gleaned from the environment so
	// let's set it so we can inspect
	expectedFrom := "4045551234"
	os.Setenv("TW_FROM", expectedFrom)

	// MockAK definition
	var mockAK = new(MockAK)
	mockAK.MockGetCurrentSubscribers = func() (CampaignTargets, error) {
		return CampaignTargets{
			{
				ID:            1,
				PhoneNumber:   "+15551234567",
				ZipCode:       "55555",
				Plus4:         "1234",
				USDistrict:    "GA_06",
				StateDistrict: "",
				USCounty:      "fulton",
			},
		}, nil
	}
	mockAK.MockGuessDistrictForZip = func(zip string) (string, error) {
		return "GA_06", nil
	}
	mockAK.MockGetRepresentatives = func() (Representatives, error) {
		return Representatives{
			"GA_06": {
				Title:        "Rep.",
				LongTitle:    "Representative",
				USDistrict:   "GA_06",
				FirstName:    "John",
				LastName:     "Ossof",
				OfficialName: "John Ossof",
				PhoneNumber:  "5551234567",
			},
		}, nil
	}

	// Mock client definition
	var mockClient = new(MockTwilioClient)
	var capturedMsg *Message
	mockClient.MockSendSMSMethod = func(msg *Message) (int, error) {
		capturedMsg = msg
		return http.StatusCreated, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/campaigns/test",
		"/campaigns/test",
		bytes.NewBufferString(testCampaign),
		TestCampaign(mockClient, mockAK, NewTagMerger()),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if mockClient.SendSMSCallCount != 1 {
				context.Errorf("Unexpected interaction: Call cound should be %v got %v",
					1, mockClient.SendSMSCallCount)
			}
			if code := recorder.Code; code != http.StatusCreated {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					code, http.StatusCreated)
			}
			if capturedMsg.To != "7067815146" {
				context.Errorf("Expected a cleaned phone number. got %v want %v", capturedMsg.To, "7067815146")
			}
		},
	}
	Run(context, theTest)
}

func Test_SendTestMessageWithBadJSON(context *testing.T) {
	var mockClient = new(MockTwilioClient)
	var mockAK = new(MockAK)
	theTest := TestParameters{
		http.MethodPost,
		"/campaigns/test",
		"/campaigns/test",
		bytes.NewBufferString(`{ {{ : d34 ;lks }}}`),
		TestCampaign(mockClient, mockAK, NewTagMerger()),
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

func Test_SendTestMessageHandlesTwilioErrors(context *testing.T) {
	expectedStatus := 403
	exception := TwilioException{
		Status:   expectedStatus,
		Message:  "Bad juju",
		Code:     9999,
		MoreInfo: "Badder juju",
	}

	// Mock a valid client.
	var mockClient = new(MockTwilioClient)
	mockClient.MockSendSMSMethod = func(message *Message) (int, error) {
		return expectedStatus, exception
	}

	// Mock up a valid AK
	var mockAK = new(MockAK)
	mockAK.MockGetCurrentSubscribers = func() (CampaignTargets, error) {
		return CampaignTargets{
			{
				ID:            1,
				PhoneNumber:   "5551234567",
				ZipCode:       "55555",
				Plus4:         "1234",
				USDistrict:    "GA_06",
				StateDistrict: "",
				USCounty:      "fulton",
			},
		}, nil
	}
	mockAK.MockGuessDistrictForZip = func(zip string) (string, error) {
		return "GA_06", nil
	}
	mockAK.MockGetRepresentatives = func() (Representatives, error) {
		return Representatives{
			"GA_06": {
				Title:        "Rep.",
				LongTitle:    "Representative",
				USDistrict:   "GA_06",
				FirstName:    "John",
				LastName:     "Ossaf",
				OfficialName: "John Ossaf",
				PhoneNumber:  "5551234567",
			},
		}, nil
	}

	theTest := TestParameters{
		http.MethodPost,
		"/campaigns/test",
		"/campaigns/test",
		bytes.NewBufferString(personOfInterest),
		TestCampaign(mockClient, mockAK, NewTagMerger()),
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
