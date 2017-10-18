package texter

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func Test_APIIsCreatedFromEnvAndSecure(context *testing.T) {
	expectedSecret := "ABC123"
	os.Setenv("RRN_API_SECRET", expectedSecret)
	os.Setenv("RRN_API_URL", "nowhere")
	api := NewAPI()

	if ok, _ := api.IsSecured(); !ok {
		context.Errorf("API should be secured!")
	}
}

func Test_APISendsAuthorization(context *testing.T) {
	expectedSecret := "ABC123"
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		response.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("[]")))
		return response, nil
	}
	api := APIV1{Secret: expectedSecret, Client: mockClient, BaseURL: "nowhere"}
	_, err := api.GetAllApprovers()

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if auth := capturedRequest.Header.Get("Authorization"); auth != "Bearer "+expectedSecret {
		context.Errorf("Not communicating securely! Authorization header [%v]", auth)
	}

	if capturedRequest.URL.String() != "nowhere/approvers" {
		context.Errorf("Not pointing to the right place: %v", capturedRequest.URL.String())
	}
}

func Test_APIGetsApprovers(context *testing.T) {
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		response.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(`[
			{ "name": "jduv", "phoneNumber":"+17067815146", "email":"jduv@hello.com" }
		]`)))
		return response, nil
	}
	api := APIV1{Client: mockClient, BaseURL: "nowhere"}
	approvers, err := api.GetAllApprovers()

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if len(approvers) != 1 {
		context.Errorf("Approvers list not correct, wanted %v elements got %v", 1, len(approvers))
	}

	if number := approvers[0].PhoneNumber; number != "7067815146" {
		context.Errorf("Approvers phone number was not cleaned. want %v got %v", "7067815146", number)
	}
}

func Test_UnsubscribeSendsRequest(context *testing.T) {
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		return response, nil
	}

	api := APIV1{Client: mockClient, BaseURL: "nowhere"}
	err := api.Unsubscribe([]string{"email1@dude.com", "email2@dude.com"})

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if mockClient.DoMethodCallCount != 2 {
		context.Errorf("Expected %v requests, got %v", 2, mockClient.DoMethodCallCount)
	}
}

func Test_UnsubscribeSendsNoRequestsWithEmptySlice(context *testing.T) {
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		return response, nil
	}

	api := APIV1{Client: mockClient, BaseURL: "nowhere"}
	err := api.Unsubscribe([]string{})

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if mockClient.DoMethodCallCount != 0 {
		context.Errorf("Expected %v requests, got %v", 0, mockClient.DoMethodCallCount)
	}
}

func Test_SubscribeSendsRequest(context *testing.T) {
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		return response, nil
	}

	api := APIV1{Client: mockClient, BaseURL: "nowhere"}
	err := api.Subscribe([]string{"email1@dude.com", "email2@dude.com"})

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if mockClient.DoMethodCallCount != 2 {
		context.Errorf("Expected %v requests, got %v", 2, mockClient.DoMethodCallCount)
	}
}

func Test_SubscribeSendsNoRequestsWithEmptySlice(context *testing.T) {
	var capturedRequest *http.Request
	mockClient := new(MockHTTPClient)
	mockClient.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedRequest = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		return response, nil
	}

	api := APIV1{Client: mockClient, BaseURL: "nowhere"}
	err := api.Subscribe([]string{})

	if err != nil {
		context.Errorf("Got an unexpected error %v", err)
	}

	if mockClient.DoMethodCallCount != 0 {
		context.Errorf("Expected %v requests, got %v", 0, mockClient.DoMethodCallCount)
	}
}
