package texter

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func Test_CleanPhoneNumber(context *testing.T) {
	var table = []struct {
		in  string
		out string
	}{
		{`ABC`, `ABC`},
		{`1-222-333-4444`, `2223334444`},
		{`+1-222-333-4444`, `2223334444`},
		{`222-333-4444`, `2223334444`},
		{`1 (404) 456 5555`, `4044565555`},
		{`555 5678`, `5555678`},
		{`555 123 1234 56`, `555123123456`},
		{`1 555 555 1234 1234`, `155555512341234`},
	}

	for _, test := range table {
		number := CleanPhoneNumber(test.in)
		if number != test.out {
			context.Errorf("CleanPhoneNumber(%q) => %q, want %q", test.in, number, test.out)
		}
	}
}

func Test_URLEncode(context *testing.T) {
	var table = []struct {
		in  Message
		out string
	}{
		{
			Message{
				To:   "4045556789",
				From: "7065551234",
				Body: "Hellow World"},
			`Body=Hellow+World&From=7065551234&To=4045556789`,
		},
		{
			Message{
				To:   "",
				From: "7065551234",
				Body: "Hellow World"},
			`Body=Hellow+World&From=7065551234&To=`,
		},
		{
			Message{
				To:   "",
				From: "",
				Body: "Hellow World"},
			`Body=Hellow+World&From=&To=`,
		},
		{
			Message{},
			`Body=&From=&To=`,
		},
		{
			Message{
				To:   "4045556789",
				From: "7065551234",
				Body: "Hellow World∂ƒ†∆˚˚ˆ√∂®¬˚"},
			`Body=Hellow+World%E2%88%82%C6%92%E2%80%A0%E2%88%86%CB%9A%CB%9A%CB%86%E2%88%9A%E2%88%82%C2%AE%C2%AC%CB%9A&From=7065551234&To=4045556789`,
		},
	}

	for _, test := range table {
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(test.in.URLEncode())

		if buffer.String() != test.out {
			context.Errorf("UrlEncode(%q) => %q, want %q", test.in, buffer.String(), test.out)
		}
	}
}

func Test_ExceptionAsErrorReturnsExpected(context *testing.T) {
	exception := TwilioException{
		Status:   500,
		Message:  "Bad juju",
		Code:     9999,
		MoreInfo: "Badder juju",
	}

	err := error(exception)
	if err.Error() != exception.Message {
		context.Errorf("Exception message is wrong. Expected %s got %s", exception.Message, err.Error())
	}
}

func Test_NewTWReturns(context *testing.T) {
	expectedAcct := "abc123"
	expectedTkn := "tokenbooyah"
	expectedURL := "https://api.twilio.com/2010-04-01/Accounts/abc123/Messages.json"

	os.Setenv("TW_ACCT", expectedAcct)
	os.Setenv("TW_TOKEN", expectedTkn)
	os.Setenv("TW_URL", "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json")

	twilio := NewTW()

	if twilio.AccountSid != expectedAcct {
		context.Errorf("AccountSID is wrong. Wanted %s got %s", expectedAcct, twilio.AccountSid)
	}

	if twilio.AuthToken != expectedTkn {
		context.Errorf("AuthToken is wrong. Wanted %s got %s", expectedTkn, twilio.AuthToken)
	}

	if twilio.BaseURL != expectedURL {
		context.Errorf("ExpectedURL is wrong. Wanted %s got %s", expectedURL, twilio.BaseURL)
	}

	if twilio.Client == nil {
		context.Errorf("Client is nil.")
	}
}

func Test_SendSMSReturnsTheCorrectTuple(context *testing.T) {
	client := new(MockHTTPClient)
	var capturedReq *http.Request
	client.MockDoMethod = func(request *http.Request) (*http.Response, error) {
		capturedReq = request
		response := new(http.Response)
		response.StatusCode = http.StatusOK
		response.Body = ioutil.NopCloser(bytes.NewReader([]byte("Testing")))
		return response, nil
	}

	expectedURL := "https://api.twilio.com/2010-04-01/Accounts/ABCDEF/Messages.json"
	twilio := Twilio{
		AccountSid: "ABCDEF",
		AuthToken:  "123456",
		BaseURL:    expectedURL,
		Client:     client,
	}
	message := Message{To: "7065551234", From: "4045551234", Body: "Hello Test Message"}
	response, err := twilio.SendSMS(&message)

	if err != nil {
		context.Errorf("Unable to send message. Error: %s", err.Error())
	}
	if response != http.StatusOK {
		context.Errorf("Status not what we expected. Got %v want %v", response, http.StatusOK)
	}
	if hdr := capturedReq.Header.Get("Content-type"); hdr != "application/x-www-form-urlencoded" {
		context.Errorf("Content-Type is wrong. got %v want %v", hdr, "application/x-www-form-urlencoded")
	}
	if hdr := capturedReq.Header.Get("Accept"); hdr != "application/json" {
		context.Errorf("Accept header is wrong. got %v want %v", hdr, "application/json")
	}
	if capturedReq.Method != "POST" {
		context.Errorf("Method is wrong. got %v want %v", capturedReq.Method, "POST")
	}
	if capturedReq.URL.String() != expectedURL {
		context.Errorf("URL is wrong. got %v want %v", capturedReq.URL, expectedURL)
	}
}

func Test_BadPhoneNumberReturnsAnError(context *testing.T) {
	client := new(MockHTTPClient)
	twilio := Twilio{Client: client}
	message := Message{To: "BadPhoneNumber", From: "4045551234", Body: "Hello Test Message"}
	response, err := twilio.SendSMS(&message)

	if err == nil {
		context.Errorf("Bad Phone number should return an error!")
	}

	if response != http.StatusBadRequest {
		context.Errorf("Status not what we expected. Got %v want %v", response, http.StatusBadRequest)
	}
}

// Tests internal machinery not exported to other packages.
func Test_handleTwilioResponse(context *testing.T) {
	var table = []struct {
		expectedStatus    int
		shouldReturnError bool
	}{
		{
			expectedStatus:    http.StatusOK,
			shouldReturnError: false,
		},
		{
			expectedStatus:    http.StatusCreated,
			shouldReturnError: false,
		},
		{
			expectedStatus:    http.StatusUnauthorized,
			shouldReturnError: true,
		},
		{
			expectedStatus:    http.StatusNotFound,
			shouldReturnError: true,
		},
		{
			expectedStatus:    http.StatusMethodNotAllowed,
			shouldReturnError: true,
		},
		{
			expectedStatus:    http.StatusTooManyRequests,
			shouldReturnError: true,
		},
		{
			expectedStatus:    http.StatusInternalServerError,
			shouldReturnError: true,
		},
	}

	for _, test := range table {
		response := new(http.Response)
		response.StatusCode = test.expectedStatus
		code, err := handleTwilioResponse(response)

		if code != test.expectedStatus {
			context.Errorf("Status not what we expected. Got %v want %v", code, test.expectedStatus)
		}
		if test.shouldReturnError && err == nil {
			context.Errorf("Call should have returned an error but didn't. Status %v", test.expectedStatus)
		}
	}
}

// For the special case of bad request we're going to test it separately. It requires
// a lot of additional overhead.
func Test_handleTwilioResponseBadRequest(context *testing.T) {

	badRequestException := `{"code": 21606, "message": "The From phone number 1234 is not a valid, SMS-capable inbound phone number or short code for your account.", "more_info": "https://www.twilio.com/docs/errors/21606", "status": 400}`

	response := new(http.Response)
	response.StatusCode = http.StatusBadRequest
	response.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(badRequestException)))
	code, err := handleTwilioResponse(response)

	if code != http.StatusBadRequest {
		context.Errorf("Status not what we expected. Got %v want %v", code, http.StatusBadRequest)
	}
	if err == nil {
		context.Errorf("Expected an error, got none")
	}
	if _, isok := err.(TwilioException); !isok {
		context.Errorf("Expected a twilio exception, got something else: %v", err)
	}
}

func Test_NewTwilioMessage(context *testing.T) {
	message := `ToCountry=US&ToState=CA&SmsMessageSid=SM5ab7891328b440f04b7c83b459fc85fd&NumMedia=0&ToCity=SAN+FRANCISCO&FromZip=30512&SmsSid=SM5ab7891328b440f04b7c83b459fc85fd&FromState=GA&SmsStatus=received&FromCity=BLAIRSVILLE&Body=Blah+blah&FromCountry=US&To=%2B14152001331&ToZip=94105&NumSegments=1&MessageSid=SM5ab7891328b440f04b7c83b459fc85fd&AccountSid=ACe24e29bdbc2e7664b315d7629ed3b9d9&From=%2B17067815146&ApiVersion=2010-04-01`

	values, err := url.ParseQuery(message)
	if err != nil {
		context.Fatalf("Unable to parse query string")
	}

	tw := NewTwilioMessage(values)
	if tw.AccountSid != "ACe24e29bdbc2e7664b315d7629ed3b9d9" {
		context.Errorf("Unmached field, want %v got %v", "ACe24e29bdbc2e7664b315d7629ed3b9d9", tw.AccountSid)
	}

	if tw.Body != "Blah blah" {
		context.Errorf("Unmatched field, want [%v] got [%v]", "Blah blah", tw.Body)
	}

	if tw.From != "7067815146" {
		context.Errorf("Unmatched field, want [%v] got [%v]", "7067815146", tw.From)
	}

	if tw.To != "4152001331" {
		context.Errorf("Unmatched field, want [%v] got [%v]", "4152001331", tw.To)
	}

	if tw.MessageSid != "SM5ab7891328b440f04b7c83b459fc85fd" {
		context.Errorf("Unmatched field, want [%v] got [%v]", "SM5ab7891328b440f04b7c83b459fc85fd", tw.MessageSid)
	}
}
