package texter

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// CampaignWorkerPool is a mapping of campaign worker routines to
// campaign object IDs
type CampaignWorkerPool map[string]*CampaignWorker

// CampaignManager handles starting and stopping campaigns
type CampaignManager interface {
	Start(campaign Campaign) error
	Stop(id string)
}

// CM implements the campaign manager interface
type CM struct {
	pool CampaignWorkerPool
	ak   AK
	db   DB
	api  RRNAPI
}

// NewCM returns a new CM interface
func NewCM(ak AK, db DB, api RRNAPI) CampaignManager {
	return &CM{
		ak:   ak,
		db:   db,
		api:  api,
		pool: make(CampaignWorkerPool),
	}
}

// Start will boot a campaign
func (c *CM) Start(campaign Campaign) error {
	worker, err := c.findOrCreateWorker(campaign)
	if err != nil {
		campaignErr := c.api.CampaignCompleted(APIStatusTerminated, campaign)
		c.pool[campaign.ID.Hex()] = nil
		return fmt.Errorf("Error attempting to spin up a campaign. Error [%v] with nested error [%v]",
			err, campaignErr)
	}

	switch worker.state.Status {
	case StatusRunning:
		log.Print("Campaign is currently running, will do nothing.")
	case StatusStopped:
		log.Printf("Found a stopped campaign, starting it")
		worker.Run()
	case StatusDone:
		log.Print("Campaign is currently done, we cannot restart it.")
	case StatusInitialized:
		c.pool[campaign.ID.Hex()] = worker
		worker.Run()
	}

	return nil
}

// Stop will turn off a campaign
func (c *CM) Stop(objectID string) {
	worker, ok := c.pool[objectID]
	if ok && worker.state.Status == StatusRunning {
		log.Printf("Found a running campaign, sending quit signal")
		worker.quit <- true

		// attempt to save state to the DB. If we can't keep the worker in
		// memory so we will run it again.
		err := c.db.SaveCampaignState(worker.state)
		if err != nil {
			log.Printf("Unable to save state %v to DB due to error %v", worker.state, err)
		}
	}
}

// CampaignWorker represents a structure that will
// run campaigns
type CampaignWorker struct {
	reps  Representatives
	state CampaignState
	quit  chan bool
	db    DB
	api   RRNAPI
}

func (c *CM) findOrCreateWorker(campaign Campaign) (*CampaignWorker, error) {
	log.Printf("Attempting to retrieve state from local cache...")
	worker, ok := c.pool[campaign.ID.Hex()]
	if ok {
		return worker, nil
	}

	log.Printf("Attempting to retrieve state for campaign %v from DB", campaign.ID)
	state, err := c.db.GetCampaignStateByID(campaign.ID)
	if err == nil {
		// Need to grab the reps
		reps, err := c.ak.GetRepresentatives()
		if err != nil {
			return nil, err
		}

		return &CampaignWorker{
			reps:  reps,
			state: state,
			quit:  make(chan bool),
		}, nil
	}

	// Need a new one.
	log.Printf("Booting a new campaign")
	audience, err := c.determineAudience(campaign)
	if err != nil {
		return nil, err
	}

	reps, err := c.ak.GetRepresentatives()
	if err != nil {
		return nil, err
	}

	return &CampaignWorker{
		reps: reps,
		state: CampaignState{
			ID:              bson.NewObjectId(),
			Campaign:        campaign,
			Status:          StatusInitialized,
			AudiencePointer: 0,
			Audience:        audience,
			Stats: CampaignStats{
				SentTo:          0,
				TotalRecipients: len(audience),
				Errors:          0,
			},
		},
		quit: make(chan bool),
		db:   c.db,
		api:  c.api,
	}, nil
}

func (c *CM) determineAudience(campaign Campaign) (CampaignTargets, error) {
	if campaign.CreatedFrom != nil && campaign.CreatedFrom.Valid() {
		log.Printf("This campaign was created from %v", campaign.CreatedFrom)
		original, err := c.db.GetCampaignStateByID(*campaign.CreatedFrom)
		if err != nil {
			log.Printf("Unable to retrieve the original campaign. Error %v", err)
			return nil, err
		}

		switch campaign.AudienceType {
		case AudienceReceived:
			log.Print("Sending to the people who originally received the campaign.")
			return original.Audience[:original.AudiencePointer], nil
		case AudienceNotReceived:
			log.Print("Sending to the people who originally did not receive the campaign.")
			return original.Audience[original.AudiencePointer+1:], nil
		case AudienceAll:
			log.Print("Sending to the original audience for the campaign.")
			return original.Audience, nil
		}
	}

	log.Print("Going to AK for all subscribers")
	return c.ak.GetCurrentSubscribers()
}

// Run executes a campaign
func (w *CampaignWorker) Run() {
	throttle, err := strconv.Atoi(os.Getenv("CM_TICKINTERVAL"))
	if err != nil {
		throttle = 300
		log.Printf("Using default value for throttle: %v", throttle)
	}

	// The from number always lives in the environment
	from := os.Getenv("TW_FROM")

	// Figure out some variables
	m := NewTagMerger()
	t := NewTW()

	go func(w *CampaignWorker) {
		log.Printf("Worker booting...")
		w.state.Status = StatusRunning
		log.Printf("Starting at index %v", w.state.AudiencePointer)

		for i := w.state.AudiencePointer; i < len(w.state.Audience); i++ {
			a := w.state.Audience[i]

			select {
			case <-w.quit:
				log.Printf("Got quit signal, saving audience pointer to %v", w.state.AudiencePointer)
				w.state.Status = StatusStopped
				return
			default:
				rep, ok := w.reps[a.USDistrict]
				if !ok {
					log.Printf("Unable to lookup district for user %v", a)
					continue
				}

				mergedBody := m.MergeRep(w.state.Campaign.Body, rep)

				msg := Message{
					To:   a.PhoneNumber,
					From: from,
					Body: mergedBody}

				status, smsErr := t.SendSMS(&msg)
				if smsErr != nil {
					log.Printf("Error sending message %v got status %v", msg, status)
					w.state.Stats.Errors++
				} else {
					w.state.Stats.SentTo++
				}

				w.state.AudiencePointer++
			}

			time.Sleep(time.Duration(throttle) * time.Millisecond)
		}

		log.Printf("Campaign has finished, cleaning up")
		err := w.db.SaveCampaignState(w.state)
		if err != nil {
			w.state.Status = StatusCompletedWithErrors
		} else {
			w.state.Status = StatusDone
		}

		err = w.api.CampaignCompleted(APIStatusCompleted, w.state.Campaign)
		if err != nil {
			log.Printf("Error attempting to modify the status of a campaign %v", err)
		}

		log.Printf("Stats %v", w.state.Stats)
	}(w)
}
