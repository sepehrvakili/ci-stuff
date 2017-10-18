package texter

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// ApprovalRequest represents a campaign to be approved
// along with all of it's approvers.
type ApprovalRequest struct {
	ID          bson.ObjectId `json:"_id" bson:"_id"`
	CampaignID  int           `json:"cid" bson:"campaignID"`
	Body        string        `json:"body" bson:"body"`
	Version     int           `json:"version" bson:"version"`
	RequestedAt time.Time     `json:"requestedAt" bson:"requestedAt"`
	ApprovedOn  *time.Time    `json:"approvedOn" bson:"approvedOn"`
	Approvers   []Approver    `json:"approvers" bson:"approvers"`
}

// ApprovalResponse represents an approval we send back to the API
type ApprovalResponse struct {
	Status    int      `json:"status"`
	Approvers []string `json:"approved_by"`
}

// StatusChangeRequest is sent to the API when we want to move the status of
// a Campaign
type StatusChangeRequest struct {
	Status int `json:"status"`
}

// Approver is someone who is going to approve a campaign.
// Looks remarkably familiar. It's here for extensibility
type Approver struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string `json:"name" bson:"name"`
	PhoneNumber string `json:"phoneNumber" bson:"phoneNumber"`
	Email       string `json:"email" bson:"email"`
	HasApproved bool   `json:"approved" bson:"approved"`
}

// Recipient is someone who is going to receive a text.
type Recipient struct {
	Name        string
	PhoneNumber string
	Email       string
}

// MessageTemplate represents a message template that lives inside our
// database. These are used for non-campaign reasons.
type MessageTemplate struct {
	ID     string `json:"_id"`
	Key    string `json:"key"`
	Body   string `json:"body"`
	Active bool   `json:"active"`
}

// TestMessage represents someone wishing to test a particular campaign
type TestMessage struct {
	PhoneNumber string `json:"phoneNumber"`
	ZipCode     string `json:"zipCode"`
	Body        string `json:"body"`
}

// CampaignTarget is an internal representation of someone we're going to
// send a campaign to.
type CampaignTarget struct {
	ID            int
	PhoneNumber   string
	ZipCode       string
	Plus4         string
	USDistrict    string
	StateDistrict string
	USCounty      string
}

// CampaignTargets is a simple typedef on a slice of CampaignTarget
// objects. Used for convenience
type CampaignTargets []CampaignTarget

// RepInfo provides information about a representative. Used for merging
// campaign information.
type RepInfo struct {
	Title        string
	LongTitle    string
	USDistrict   string
	FirstName    string
	LastName     string
	OfficialName string
	PhoneNumber  string
}

// Representatives is simply a map of district to RepInfo objects.
type Representatives map[string]RepInfo

const (
	// StatusInitialized status means we haven't done anything yet (default)
	StatusInitialized = iota
	// StatusRunning status means the campaign is running
	StatusRunning
	// StatusStopped status means the campaign is stopped
	StatusStopped
	// StatusDone means the campaign is done
	StatusDone
	// StatusCompletedWithErrors means the campaign has finished but is in an error state
	StatusCompletedWithErrors
)

const (
	// AudienceReceived will send a campaign to all who received the cloned campaign
	AudienceReceived = 1
	// AudienceNotReceived will send a campaign to all who did not receive the cloned campaign
	AudienceNotReceived = 2
	// AudienceAll will send a campaign to everyone in the cloned campaign
	AudienceAll = 3
)

// Campaign represents a campaign to be started
type Campaign struct {
	ID           bson.ObjectId  `json:"_id" bson:"_id"`
	CampaignID   int            `json:"cid" bson:"campaignID"`
	CreatedFrom  *bson.ObjectId `json:"cloned_from" bson:"createdFrom"`
	AudienceType int            `json:"cloned_recipient_group" bson:"audience"`
	Body         string         `json:"message" bson:"body"`
}

// CampaignStats represents stastistics for a certain campaign
type CampaignStats struct {
	SentTo          int `json:"sentTo" bson:"sentTo"`
	TotalRecipients int `json:"totalRecipients" bson:"totalRecipients"`
	Errors          int `json:"errors" bson:"errors"`
}

// CampaignState holds information about where a campaign is in it's lifecycle
type CampaignState struct {
	ID              bson.ObjectId   `json:"_id" bson:"_id"`
	Campaign        Campaign        `bson:"campaign"`
	Status          int             `bson:"status"`
	Stats           CampaignStats   `bson:"stats"`
	AudiencePointer int             `bson:"audiencePointer"`
	Audience        CampaignTargets `bson:"audience"`
}
