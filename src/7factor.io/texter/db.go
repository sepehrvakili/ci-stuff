package texter

import (
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DB interface abstracts the storage mechanism of our application. We
// don't care what kind of DB backs the calls, just so long as we get
// the objects we expect.
type DB interface {
	GetActiveMessageTemplates() ([]MessageTemplate, error)
	GetMessageTemplateByKey(key string) (MessageTemplate, error)
	RegisterApproval(approval ApprovalRequest) error
	GetMostRecentApprovalByID(cid int) (ApprovalRequest, error)
	ApproveCampaign(request ApprovalRequest, approver Approver) error
	SaveCampaignState(toSave CampaignState) error
	GetCampaignStateByID(campaignID bson.ObjectId) (CampaignState, error)
	StowMessage(msg TwilioMessage) error
	Close()
}

// MongoDB manages connections and queries to Mongo for our
// application.
type MongoDB struct {
	Creds   DBCreds
	Session *mgo.Session
}

// NewMongoDB returns a database object that's ready to be queried. Pass
// the appropriate credentials in.
func NewMongoDB(creds DBCreds) (DB, error) {
	info := creds.ToDialInfo()
	info.Timeout = 30 * time.Second
	log.Printf("connecting to mongo server at %v", info.Addrs)
	session, err := mgo.DialWithInfo(info)
	db := MongoDB{Creds: creds, Session: session}
	return db, err
}

// GetActiveMessageTemplates returns a list of all currently active message templates.
func (db MongoDB) GetActiveMessageTemplates() ([]MessageTemplate, error) {
	var templates []MessageTemplate
	collection := db.Session.DB(db.Creds.DBName).C("MessageTemplates")
	err := collection.Find(bson.M{"active": true}).All(&templates)
	return templates, err
}

// GetMessageTemplateByKey returns a message template given a key.
func (db MongoDB) GetMessageTemplateByKey(key string) (MessageTemplate, error) {
	var template MessageTemplate
	collection := db.Session.DB(db.Creds.DBName).C("MessageTemplates")
	err := collection.Find(bson.M{"key": key}).One(&template)
	return template, err
}

// RegisterApproval will associate an approver with a given campaign
func (db MongoDB) RegisterApproval(approval ApprovalRequest) error {
	collection := db.Session.DB(db.Creds.DBName).C("ApprovalRequests")
	approval.RequestedAt = time.Now()
	approval.ApprovedOn = nil // just in case.

	for _, approver := range approval.Approvers {
		approver.HasApproved = false
	}

	_, err := collection.Upsert(bson.M{"_id": approval.ID}, approval)
	return err
}

// GetMostRecentApprovalByID gets an approval request object based on the campaign ID and the most
// recent version number.
func (db MongoDB) GetMostRecentApprovalByID(cid int) (ApprovalRequest, error) {
	collection := db.Session.DB(db.Creds.DBName).C("ApprovalRequests")
	var request ApprovalRequest
	err := collection.Find(bson.M{"campaignID": cid}).Sort("-version").One(&request)
	return request, err
}

// ApproveCampaign approves a campaign in our DB for audit purposes. The approver passed in
// will be matched to a phone number existing inside the request. If there exists no match, then
// the approver will be appended to the list.
func (db MongoDB) ApproveCampaign(request ApprovalRequest, approver Approver) error {
	now := time.Now()
	request.ApprovedOn = &now

	for i, current := range request.Approvers {
		if approver.PhoneNumber == current.PhoneNumber {
			request.Approvers[i].HasApproved = true
			return db.updateApprovalRequest(request)
		}
	}

	approver.HasApproved = true
	request.Approvers = append(request.Approvers, approver)
	return db.updateApprovalRequest(request)
}

// SaveCampaignState saves a CampaignState object to the DB
func (db MongoDB) SaveCampaignState(toSave CampaignState) error {
	collection := db.Session.DB(db.Creds.DBName).C("Campaigns")
	_, err := collection.UpsertId(toSave.ID, toSave)
	return err
}

// GetCampaignStateByID retrieves the state for a given campaign ID, which should never
// ever be duplicated.
func (db MongoDB) GetCampaignStateByID(campaignID bson.ObjectId) (CampaignState, error) {
	collection := db.Session.DB(db.Creds.DBName).C("Campaigns")
	var state CampaignState
	err := collection.Find(bson.M{"campaign._id": campaignID}).One(&state)
	return state, err
}

// StowMessage saves twilio messages to a LostAndFound table
func (db MongoDB) StowMessage(msg TwilioMessage) error {
	collection := db.Session.DB(db.Creds.DBName).C("LostAndFound")
	return collection.Insert(msg)
}

// Helper function for updating approvals in the mongo db
func (db MongoDB) updateApprovalRequest(toUpdate ApprovalRequest) error {
	collection := db.Session.DB(db.Creds.DBName).C("ApprovalRequests")
	return collection.Update(bson.M{"campaignID": toUpdate.CampaignID, "version": toUpdate.Version}, toUpdate)
}

// Close message shuts down the DB handle.
func (db MongoDB) Close() {
	db.Session.Close()
}
