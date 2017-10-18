package texter

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_IncomingProcessesMessage(context *testing.T) {
	mockAPI := new(MockAPI)
	mockDB := new(MockDB)
	mockAK := new(MockAK)
	theTest := TestParameters{
		http.MethodGet,
		"/incoming",
		"/incoming",
		nil,
		Incoming(mockAPI, mockDB, mockAK),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusCreated {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusCreated)
			}
		},
	}
	Run(context, theTest)
}

func Test_IncomingApprovesCampaign(context *testing.T) {
	message := `ToCountry=US&ToState=CA&SmsMessageSid=SM5ab7891328b440f04b7c83b459fc85fd&NumMedia=0&ToCity=SAN+FRANCISCO&FromZip=30512&SmsSid=SM5ab7891328b440f04b7c83b459fc85fd&FromState=GA&SmsStatus=received&FromCity=BLAIRSVILLE&Body=42&FromCountry=US&To=%2B14152001331&ToZip=94105&NumSegments=1&MessageSid=SM5ab7891328b440f04b7c83b459fc85fd&AccountSid=ACe24e29bdbc2e7664b315d7629ed3b9d9&From=%2B17067815146&ApiVersion=2010-04-01`

	approvers := []Approver{{
		Name:        "jduv",
		PhoneNumber: "4045551234",
		Email:       "jduv@7factor.io",
	}}

	var capturedID int
	mockDB := new(MockDB)
	mockDB.MockGetMostRecentApprovalByID = func(cid int) (ApprovalRequest, error) {
		capturedID = cid
		return ApprovalRequest{
			CampaignID: 42,
			Version:    1,
			Approvers:  approvers,
		}, nil
	}
	mockDB.MockApproveCampaign = func(request ApprovalRequest, approver Approver) error {
		return nil
	}

	mockAPI := new(MockAPI)
	mockAPI.MockGetAllApprovers = func() ([]Approver, error) {
		return approvers, nil
	}
	mockAPI.MockApproveCampaign = func(approval ApprovalRequest, approver Approver) error {
		return nil
	}

	mockAK := new(MockAK)

	theTest := TestParameters{
		http.MethodGet,
		"/incoming",
		"/incoming",
		bytes.NewBufferString(message),
		Incoming(mockAPI, mockDB, mockAK),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusCreated {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusCreated)
			}

			if capturedID != 42 {
				context.Errorf("Parsed an incorrect campaign ID: got %v want 42", capturedID)
			}
		},
	}
	Run(context, theTest)
}

func Test_IncomingApprovalFallsBackIfAPIBreaks(context *testing.T) {
	message := `ToCountry=US&ToState=CA&SmsMessageSid=SM5ab7891328b440f04b7c83b459fc85fd&NumMedia=0&ToCity=SAN+FRANCISCO&FromZip=30512&SmsSid=SM5ab7891328b440f04b7c83b459fc85fd&FromState=GA&SmsStatus=received&FromCity=BLAIRSVILLE&Body=42&FromCountry=US&To=%2B14152001331&ToZip=94105&NumSegments=1&MessageSid=SM5ab7891328b440f04b7c83b459fc85fd&AccountSid=ACe24e29bdbc2e7664b315d7629ed3b9d9&From=%2B17067815146&ApiVersion=2010-04-01`

	approvers := []Approver{{
		Name:        "jduv",
		PhoneNumber: "4045551234",
		Email:       "jduv@7factor.io",
	}}

	var capturedID int
	mockDB := new(MockDB)
	mockDB.MockGetMostRecentApprovalByID = func(cid int) (ApprovalRequest, error) {
		capturedID = cid
		return ApprovalRequest{
			CampaignID: 42,
			Version:    1,
			Approvers:  approvers,
		}, nil
	}

	mockDB.MockApproveCampaign = func(request ApprovalRequest, approver Approver) error {
		return nil
	}

	mockAPI := new(MockAPI)
	mockAPI.MockGetAllApprovers = func() ([]Approver, error) {
		return nil, errors.New("busted")
	}
	mockAPI.MockApproveCampaign = func(approval ApprovalRequest, approver Approver) error {
		return errors.New("double busted")
	}

	mockAK := new(MockAK)

	theTest := TestParameters{
		http.MethodGet,
		"/incoming",
		"/incoming",
		bytes.NewBufferString(message),
		Incoming(mockAPI, mockDB, mockAK),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusCreated {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusCreated)
			}

			if capturedID != 42 {
				context.Errorf("Parsed an incorrect campaign ID: got %v want 42", capturedID)
			}
		},
	}
	Run(context, theTest)
}

func Test_IncomingHandlesUnsubscribe(context *testing.T) {
	table := []struct {
		message    string
		shouldStop bool
	}{
		{
			message:    `Body=%20%20%20STOP%20%20`,
			shouldStop: true,
		}, {
			message:    `Body=%20%20%20STOPALL%20%20`,
			shouldStop: true,
		}, {
			message: `Body=%20%20%20quiT%20%20`,
		}, {
			message:    `Body=%20%20%20cancel%20%20`,
			shouldStop: true,
		}, {
			message:    `Body=%20%20%20enD%20%20`,
			shouldStop: true,
		}, {
			message:    `Body=%20%20%20unSubScribe%20%20`,
			shouldStop: true,
		}, {
			message:    `Body=%20%20%20%20%20%20STOP%20.%20%20%20%20Please`,
			shouldStop: false,
		}, {
			message:    `Body=%20%20STOPAlL%20dammit%20%20%20`,
			shouldStop: false,
		}, {
			message:    `Body=%20ragequIT%20%20%20%20`,
			shouldStop: false,
		}, {
			message:    `Body=CANceL.%20%20`,
			shouldStop: false,
		}, {
			message:    `Body=%20%20EnD-it%20`,
			shouldStop: false,
		}, {
			message:    `Body=%20%20unsubscribe%20%20me%20%20`,
			shouldStop: false,
		},
	}

	for _, params := range table {

		mockDB := new(MockDB)
		mockDB.MockStowMessage = func(msg TwilioMessage) error {
			return nil
		}
		mockAPI := new(MockAPI)
		mockAPI.MockUnsubscribe = func(email []string) error {
			return nil
		}

		mockAK := new(MockAK)
		mockAK.MockGetEmailsForPhoneNumber = func(phoneNumber string) ([]string, error) {
			return []string{"wat@wat.com"}, nil
		}

		theTest := TestParameters{
			http.MethodGet,
			"/incoming",
			"/incoming",
			bytes.NewBufferString(params.message),
			Incoming(mockAPI, mockDB, mockAK),
			func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
				if status := recorder.Code; status != http.StatusCreated {
					context.Errorf("Handler returned an incorrect status code: got %v want %v",
						status, http.StatusCreated)
				}

				var times int
				if params.shouldStop {
					times = 1
				} else {
					times = 0
				}

				if params.shouldStop && mockAPI.UnsubscribeCallCount != times {
					context.Errorf("Should have called unsubscribe exactly %v times, got %v", times, mockAPI.UnsubscribeCallCount)
				}
			},
		}
		Run(context, theTest)

	}
}

func Test_IncomingHandlesSubscribe(context *testing.T) {
	table := []struct {
		message     string
		shouldStart bool
	}{
		{
			message:     `Body=%20%20%20%20StarT%20%20`,
			shouldStart: true,
		}, {
			message:     `Body=%20%20%20%20UnStoP%20%20`,
			shouldStart: true,
		}, {
			message:     `Body=yes`,
			shouldStart: true,
		}, {
			message:     `Body=Start%20me%20up`,
			shouldStart: false,
		}, {
			message:     `Body=%20%20I%20want%20to%20unstop%20%20`,
			shouldStart: false,
		}, {
			message:     `Body=%20%20%20%20yeslets%20%20`,
			shouldStart: false,
		},
	}

	for _, params := range table {

		mockDB := new(MockDB)
		mockDB.MockStowMessage = func(msg TwilioMessage) error {
			return nil
		}
		mockAPI := new(MockAPI)
		mockAPI.MockSubscribe = func(email []string) error {
			return nil
		}

		mockAK := new(MockAK)
		mockAK.MockGetEmailsForPhoneNumber = func(phoneNumber string) ([]string, error) {
			return []string{"wat@wat.com"}, nil
		}

		theTest := TestParameters{
			http.MethodGet,
			"/incoming",
			"/incoming",
			bytes.NewBufferString(params.message),
			Incoming(mockAPI, mockDB, mockAK),
			func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
				if status := recorder.Code; status != http.StatusCreated {
					context.Errorf("Handler returned an incorrect status code: got %v want %v",
						status, http.StatusCreated)
				}

				var times int
				if params.shouldStart {
					times = 1
				} else {
					times = 0
				}

				if params.shouldStart && mockAPI.SubscribeCallCount != times {
					context.Errorf("Should have called subscribe exactly %v times, got %v", times, mockAPI.UnsubscribeCallCount)
				}
			},
		}
		Run(context, theTest)

	}
}

func Test_StowMessage(context *testing.T) {
	table := []string{
		`Body=Who%20knows%20what%20this%20message%20is%20about%20but%20it%20has%20stop%20words`,
		`Body=I%27d%20like%20to%20write%20this%20message`,
		`Body=My%20phone%20number%20is%20555-404-1234`,
		`Body=approve%20campaign%201`,
		`Body=May%20I%20have%20this%20dance%3F`,
		`Body=%CB%99%E2%88%86%C2%A8%CB%9A%C2%AC%CB%9A%C2%AC%E2%88%82%C3%9F%E2%80%A6%E2%80%A6%C2%B4%E2%89%A4%E2%89%A4%CB%9C%CB%99%CB%99%C2%A8%C2%AC%CB%9C`,
		`Body=`,
	}

	for _, message := range table {
		mockAPI := new(MockAPI)
		mockAK := new(MockAK)
		mockDB := new(MockDB)
		mockDB.MockStowMessage = func(msg TwilioMessage) error {
			return nil
		}

		theTest := TestParameters{
			http.MethodGet,
			"/incoming",
			"/incoming",
			bytes.NewBufferString(message),
			Incoming(mockAPI, mockDB, mockAK),
			func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
				if status := recorder.Code; status != http.StatusCreated {
					context.Errorf("Handler returned an incorrect status code: got %v want %v",
						status, http.StatusCreated)
				}

				if mockDB.StowMessageCallCount != 1 {
					context.Errorf("Should have called stow exactly %v times, got %v", 1, mockDB.StowMessageCallCount)
				}
			},
		}
		Run(context, theTest)
	}
}

func Test_CascadeOfErrorsIsHandled(context *testing.T) {
	table := []string{
		`Body=Unsubscribe`,
		`Body=Start`,
		`Body=My%20phone%20number%20is%20555-404-1234`,
	}

	for _, message := range table {
		mockAPI := new(MockAPI)
		mockAPI.MockUnsubscribe = func(email []string) error {
			return errors.New("Essssplody")
		}
		mockAPI.MockSubscribe = func(email []string) error {
			return errors.New("Essssplody")
		}

		mockDB := new(MockDB)
		mockDB.MockStowMessage = func(msg TwilioMessage) error {
			return nil
		}

		mockAK := new(MockAK)
		mockAK.MockGetEmailsForPhoneNumber = func(phoneNumber string) ([]string, error) {
			return []string{"wat@wat.com"}, nil
		}

		theTest := TestParameters{
			http.MethodGet,
			"/incoming",
			"/incoming",
			bytes.NewBufferString(message),
			Incoming(mockAPI, mockDB, mockAK),
			func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
				if status := recorder.Code; status != http.StatusCreated {
					context.Errorf("Handler returned an incorrect status code: got %v want %v",
						status, http.StatusCreated)
				}
			},
		}
		Run(context, theTest)
	}
}
