package texter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	APIStatusReady = 2
	// APIStatusTerminated should be relayed when the campaign is in error
	APIStatusTerminated = 5
	// APIStatusCompleted should be related when the campaign completed
	APIStatusCompleted = 6
)

// RRNAPI is a simple interface back to the RRN API. We will
// use this to call the API layer
type RRNAPI interface {
	IsSecured() (bool, error)
	GetAllApprovers() ([]Approver, error)
	ApproveCampaign(approval ApprovalRequest, approver Approver) error
	Unsubscribe(email []string) error
	Subscribe(email []string) error
	CampaignCompleted(status int, campaign Campaign) error
}

// NewAPI returns a new API object used to communicate with the RRN
// API layer
func NewAPI() RRNAPI {
	secret := os.Getenv("RRN_API_SECRET")
	host := os.Getenv("RRN_API_URL")
	return APIV1{Secret: secret, BaseURL: host, Client: new(http.Client)}
}

// APIV1 implements the RRNAPI in it's first version state.
type APIV1 struct {
	Secret  string
	Client  HTTPClient
	BaseURL string
}

// IsSecured tells is if the client is ready to make calls to the API or not via
// secure access
func (api APIV1) IsSecured() (bool, error) {
	if api.Secret == "" {
		return false, errors.New("no secret configured, client is insecure")
	}

	return true, nil
}

// GetAllApprovers satisfies the interface and returns a slice of all approvers
// in the system by asking the API
func (api APIV1) GetAllApprovers() ([]Approver, error) {
	if api.Client != nil {
		requestURL := api.BaseURL + "/approvers"
		request, err := http.NewRequest(http.MethodGet, requestURL, nil)
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+api.Secret)

		// Ship the request
		response, err := api.Client.Do(request)
		if err != nil {
			return nil, err
		}

		log.Printf("%v %v %v", request.Method, requestURL, response.Status)

		if response.Body != nil {
			var approvers []Approver
			err = json.NewDecoder(response.Body).Decode(&approvers)
			if err != nil {
				bytes, _ := ioutil.ReadAll(response.Body)
				log.Printf("unable to handle message %v decoder error %v", string(bytes), err)
				return nil, err
			}

			// Clean up phone numbers for matching
			for i, approver := range approvers {
				approver.PhoneNumber = CleanPhoneNumber(approver.PhoneNumber)
				approvers[i] = approver
			}

			return approvers, nil
		}

		return []Approver{}, nil
	}

	return nil, errors.New("client is nil, unable to proceed")
}

// ApproveCampaign asks the API to approve a campaign.
func (api APIV1) ApproveCampaign(approval ApprovalRequest, approver Approver) error {
	if api.Client != nil {
		requestURL := api.BaseURL + "/campaigns/" + approval.ID.Hex()
		jsonBytes, err := json.Marshal(ApprovalResponse{Status: APIStatusReady, Approvers: []string{approver.ID}})
		if err != nil {
			return err
		}

		request, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			log.Printf("something went wrong when creating request: %v", err)
			return err
		}

		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+api.Secret)

		response, err := api.Client.Do(request)
		if err != nil {
			log.Printf("something went wrong when calling the client: %v", err)
			return err
		}

		log.Printf("%v %v %v", request.Method, requestURL, response.Status)
	}

	return nil
}

// CampaignCompleted can be used to set a campaign to a particular status
// value as understood by the API
func (api APIV1) CampaignCompleted(status int, campaign Campaign) error {
	if api.Client != nil {
		requestURL := api.BaseURL + "/campaigns/" + campaign.ID.Hex()
		jsonBytes, err := json.Marshal(StatusChangeRequest{Status: status})
		if err != nil {
			return err
		}

		request, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			log.Printf("something went wrong when creating request: %v", err)
			return err
		}

		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+api.Secret)

		response, err := api.Client.Do(request)
		if err != nil {
			log.Printf("something went wrong when calling the client: %v", err)
			return err
		}

		log.Printf("%v %v %v", request.Method, requestURL, response.Status)
	}

	return nil
}

// Unsubscribe removes a user from the AK list based on emails
func (api APIV1) Unsubscribe(email []string) error {
	if api.Client != nil {
		for _, e := range email {
			requestURL := api.BaseURL + "/users/subscriptions/" + e

			request, err := http.NewRequest(http.MethodDelete, requestURL, nil)
			if err != nil {
				log.Printf("something went wrong when creating request: %v", err)
				return err
			}

			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Authorization", "Bearer "+api.Secret)

			response, err := api.Client.Do(request)
			if err != nil {
				log.Printf("something went wrong when calling the client: %v", err)
				return err
			}

			log.Printf("%v %v %v", request.Method, requestURL, response.Status)
		}
	}

	return nil
}

// Subscribe adds a user to the AK list based on emails
func (api APIV1) Subscribe(email []string) error {
	if api.Client != nil {
		for _, e := range email {
			requestURL := api.BaseURL + "/users/subscriptions/" + e

			request, err := http.NewRequest(http.MethodPost, requestURL, nil)
			if err != nil {
				log.Printf("something went wrong when creating request: %v", err)
				return err
			}

			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Authorization", "Bearer "+api.Secret)

			response, err := api.Client.Do(request)
			if err != nil {
				log.Printf("something went wrong when calling the client: %v", err)
				return err
			}

			log.Printf("%v %v %v", request.Method, requestURL, response.Status)
		}
	}

	return nil
}
