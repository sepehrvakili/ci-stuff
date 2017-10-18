package texter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/husobee/vestigo"
)

type MockHTTPClient struct {
	DoMethodCallCount int
	MockDoMethod      func(request *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(request *http.Request) (*http.Response, error) {
	m.DoMethodCallCount++
	return m.MockDoMethod(request)
}

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

// MOCK Twilio client
type MockTwilioClient struct {
	SendSMSCallCount  int
	MockSendSMSMethod func(msg *Message) (int, error)
}

func (m *MockTwilioClient) SendSMS(msg *Message) (int, error) {
	m.SendSMSCallCount++
	return m.MockSendSMSMethod(msg)
}

// MOCK primary DB
type MockDB struct {
	GetActiveMessageTemplatesCallCount int
	GetMessageTemplateByKeyCallCount   int
	RegisterApprovalCallCount          int
	StowMessageCallCount               int
	ApproveCampaignCallCount           int
	CloseCallCount                     int
	GetMostRecentApprovalByIDCallCount int
	SaveCampaignStateCallCount         int
	GetCampaignStateCallCount          int
	MockGetActiveMessageTemplates      func() ([]MessageTemplate, error)
	MockGetMessageTemplateByKey        func(key string) (MessageTemplate, error)
	MockRegisterApproval               func(approver ApprovalRequest) error
	MockGetMostRecentApprovalByID      func(cid int) (ApprovalRequest, error)
	MockApproveCampaign                func(approval ApprovalRequest, approver Approver) error
	MockSaveCampaignState              func(toSave CampaignState) error
	MockGetCampaignStateByID           func(campaignID bson.ObjectId) (CampaignState, error)
	MockStowMessage                    func(msg TwilioMessage) error
	MockClose                          func()
}

func (db *MockDB) GetActiveMessageTemplates() ([]MessageTemplate, error) {
	db.GetActiveMessageTemplatesCallCount++
	return db.MockGetActiveMessageTemplates()
}

func (db *MockDB) GetMessageTemplateByKey(key string) (MessageTemplate, error) {
	db.GetMessageTemplateByKeyCallCount++
	return db.MockGetMessageTemplateByKey(key)
}

func (db *MockDB) RegisterApproval(approver ApprovalRequest) error {
	db.RegisterApprovalCallCount++
	return db.MockRegisterApproval(approver)
}

func (db *MockDB) GetMostRecentApprovalByID(cid int) (ApprovalRequest, error) {
	db.GetMostRecentApprovalByIDCallCount++
	return db.MockGetMostRecentApprovalByID(cid)
}

func (db *MockDB) ApproveCampaign(approval ApprovalRequest, approver Approver) error {
	db.ApproveCampaignCallCount++
	return db.MockApproveCampaign(approval, approver)
}
func (db *MockDB) GetCampaignStateByID(campaignID bson.ObjectId) (CampaignState, error) {
	db.GetCampaignStateCallCount++
	return db.MockGetCampaignStateByID(campaignID)
}

func (db *MockDB) SaveCampaignState(toSave CampaignState) error {
	db.SaveCampaignStateCallCount++
	return db.MockSaveCampaignState(toSave)
}

func (db *MockDB) StowMessage(msg TwilioMessage) error {
	db.StowMessageCallCount++
	return db.MockStowMessage(msg)
}

func (db *MockDB) Close() {
	db.CloseCallCount++
	db.MockClose()
}

// MOCK AK
type MockAK struct {
	MockGetCurrentSubscribers   func() (CampaignTargets, error)
	MockGetRepresentatives      func() (Representatives, error)
	MockGuessDistrictForZip     func(zip string) (string, error)
	MockGetEmailsForPhoneNumber func(phoneNumber string) ([]string, error)
	MockClose                   func()
}

func (ak *MockAK) GetCurrentSubscribers() (CampaignTargets, error) {
	return ak.MockGetCurrentSubscribers()
}

func (ak *MockAK) GetRepresentatives() (Representatives, error) {
	return ak.MockGetRepresentatives()
}

func (ak *MockAK) GuessDistrictForZip(zip string) (string, error) {
	return ak.MockGuessDistrictForZip(zip)
}

func (ak *MockAK) GetEmailsForPhoneNumber(phoneNumber string) ([]string, error) {
	return ak.MockGetEmailsForPhoneNumber(phoneNumber)
}

func (ak *MockAK) Close() {
	ak.MockClose()
}

// MOCK API
type MockAPI struct {
	GetAllApproversCallCount   int
	IsSecuredCallCount         int
	UnsubscribeCallCount       int
	SubscribeCallCount         int
	ApproveCampaignCallCount   int
	CampaignCompletedCallCount int
	MockGetAllApprovers        func() ([]Approver, error)
	MockIsSecured              func() (bool, error)
	MockUnsubscribe            func(email []string) error
	MockSubscribe              func(email []string) error
	MockApproveCampaign        func(approval ApprovalRequest, approver Approver) error
	MockCampaignCompleted      func(status int, campaign Campaign) error
}

func (m *MockAPI) GetAllApprovers() ([]Approver, error) {
	m.GetAllApproversCallCount++
	return m.MockGetAllApprovers()
}

func (m *MockAPI) IsSecured() (bool, error) {
	m.IsSecuredCallCount++
	return m.MockIsSecured()
}

func (m *MockAPI) ApproveCampaign(approval ApprovalRequest, approver Approver) error {
	m.ApproveCampaignCallCount++
	return m.MockApproveCampaign(approval, approver)
}

func (m *MockAPI) CampaignCompleted(status int, campaign Campaign) error {
	m.CampaignCompletedCallCount++
	return m.MockCampaignCompleted(status, campaign)
}

func (m *MockAPI) Unsubscribe(email []string) error {
	m.UnsubscribeCallCount++
	return m.MockUnsubscribe(email)
}

func (m *MockAPI) Subscribe(email []string) error {
	m.SubscribeCallCount++
	return m.MockSubscribe(email)
}
