package texter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/husobee/vestigo"
)

// Implementations of this thing should implement the assert stage of the test.
// This is always called last.
type Asserter func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request)

// Some testing helpers for making assertions and running http handler
// tests a little easier. The existing GoLang tooling is a little light
// and our stuff doesn't work great with table driven testing.
type TestParameters struct {
	method      string
	url         string
	template    string
	requestBody io.Reader
	handler     http.HandlerFunc
	assertions  Asserter
}

// Runs a test given the target parameters and asserter.
func Run(context *testing.T, toRun TestParameters) {
	request, err := http.NewRequest(toRun.method, toRun.url, toRun.requestBody)
	if err != nil {
		// This is always fatal.
		context.Fatal(err)
	}

	router := vestigo.NewRouter()
	switch toRun.method {
	case http.MethodGet:
		router.Get(toRun.template, toRun.handler)
	case http.MethodPost:
		router.Post(toRun.template, toRun.handler)
	case http.MethodDelete:
		router.Delete(toRun.template, toRun.handler)
	case http.MethodPut:
		router.Put(toRun.template, toRun.handler)
	default:
		context.Fatalf("Unknown method %v", toRun.method)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	toRun.assertions(context, recorder, request)
}

// Mock HTTP client for intercepting HTTP calls
type MockHTTPClient struct {
	MockDoMethod func(request *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(request *http.Request) (*http.Response, error) {
	return m.MockDoMethod(request)
}

// MOCK Twilio client
type MockTwilioClient struct {
	MockSendSMSMethod func(msg *Message) (int, error)
}

func (m *MockTwilioClient) SendSMS(msg *Message) (int, error) {
	return m.MockSendSMSMethod(msg)
}

// MOCK primary DB
type MockDB struct {
	MockStowMessage func(msg Message) error
	MockClose       func()
}

func (db *MockDB) StowMessage(msg Message) error {
	return db.MockStowMessage(msg)
}

func (db *MockDB) Close() {
	db.MockClose()
}

// MOCK Congress
type MockCongressDB struct {
	MockGetRepresentatives  func() (Representatives, error)
	MockGuessDistrictForZip func(zip string) (string, error)
	MockClose               func()
}

func (ak *MockCongressDB) GetRepresentatives() (Representatives, error) {
	return ak.MockGetRepresentatives()
}

func (ak *MockCongressDB) GuessDistrictForZip(zip string) (string, error) {
	return ak.MockGuessDistrictForZip(zip)
}

func (ak *MockCongressDB) Close() {
	ak.MockClose()
}
