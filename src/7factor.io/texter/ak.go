package texter

import (
	"database/sql"
	"log"
	"time"

	// for the sql drivers
	_ "github.com/go-sql-driver/mysql"
)

// AK is an interface to action kit data.
type AK interface {
	GetCurrentSubscribers() (CampaignTargets, error)
	GetRepresentatives() (Representatives, error)
	GuessDistrictForZip(zip string) (string, error)
	GetEmailsForPhoneNumber(phoneNumber string) ([]string, error)
	Close()
}

// AKDB manages connections and queries to ActionKit's MySQL for our
// application.
type AKDB struct {
	Creds   DBCreds
	Session *sql.DB
}

// NewAK returns a new ActionKit interface object.
func NewAK(creds DBCreds) (AK, error) {
	log.Printf("connecting to AK at %v", creds.Host)
	db, _ := sql.Open("mysql", creds.ToDSN())
	err := db.Ping() // make sure the DB works

	if err != nil {
		// Let's wait for a real connection, 20 retries
		for count := 1; count <= 20; count++ {
			log.Printf("retrying connection...")
			time.Sleep(5 * time.Second)
			err = db.Ping()

			if err == nil {
				log.Printf("connected to DB at URL %v", creds.Host)
				return &AKDB{Creds: creds, Session: db}, err
			}
		}

		log.Printf("unable to connect to the AK DB, giving up!")
		return nil, err
	}

	log.Printf("connected to DB at URL %v", creds.Host)
	return &AKDB{Creds: creds, Session: db}, err
}

// Gets all current subs
var currentSubsQuery = `
select 
	u.id, 
	p.normalized_phone as phonenumber,
	u.zip as zipcode, 
	u.plus4, 
	l.us_district, 
	l.us_state_district, 
	l.us_county
from core_user u 
inner join core_phone p on u.id = p.user_id
inner join core_subscription s on u.id = s.user_id
inner join core_location l on u.id = l.user_id
where list_id = 9 AND p.phone != 'NULL' AND u.subscription_status = 'subscribed'
group by p.normalized_phone;
`

// GetCurrentSubscribers satisfies the AK interface. Returns a list of
// CampaignTarget objects that you can send campaigns to.
func (ak *AKDB) GetCurrentSubscribers() (CampaignTargets, error) {
	var targets CampaignTargets
	rows, err := ak.Session.Query(currentSubsQuery)
	if err == nil {
		for rows.Next() {
			var target CampaignTarget
			err = rows.Scan(&target.ID, &target.PhoneNumber, &target.ZipCode, &target.Plus4,
				&target.USDistrict, &target.StateDistrict, &target.USCounty)

			if err != nil {
				log.Printf("GetCurrentSubscribers(): Error reading row (continuing): %v", err)
			} else {
				// Clean the number before storing the target
				target.PhoneNumber = CleanPhoneNumber(target.PhoneNumber)
				targets = append(targets, target)
			}
		}
	} else {
		log.Printf("GetRepresentatives(): Error querying AK: %v", err)
	}

	return targets, err
}

var repsQuery = `
select 
	t.title, 
	t.long_title, 
	t.us_district, 
	t.official_full, 
	t.first,
	t.last, 
	t.phone
from core_target t 
where type = 'house' and hidden = 0;
`

// GetRepresentatives satisfies the AK interface. Returns a Representative object
// that is a map of district to representative.
func (ak *AKDB) GetRepresentatives() (Representatives, error) {
	reps := make(Representatives)
	rows, err := ak.Session.Query(repsQuery)
	if err == nil {
		for rows.Next() {
			var repInfo RepInfo
			err = rows.Scan(&repInfo.Title, &repInfo.LongTitle, &repInfo.USDistrict, &repInfo.OfficialName,
				&repInfo.FirstName, &repInfo.LastName, &repInfo.PhoneNumber)

			if err != nil {
				log.Printf("GetRepresentatives(): Error reading row (continuing): %v", err)
			} else {
				// Clean the number before storing the target
				repInfo.PhoneNumber = CleanPhoneNumber(repInfo.PhoneNumber)
				reps[repInfo.USDistrict] = repInfo
			}
		}
	} else {
		log.Printf("GetRepresentatives(): Error querying AK: %v", err)
	}
	return reps, err
}

var guessDistrictQuery = `
select distinct 
	l.us_district
from core_user u 
inner join core_location l on u.id = l.user_id
where u.zip = ?;
`

// GuessDistrictForZip satisfies the AK interface. Returns a guess at the district for
// a given zip code. It won't be terribly accurate, we don't include Plus4
func (ak *AKDB) GuessDistrictForZip(zip string) (string, error) {
	statement, err := ak.Session.Prepare(guessDistrictQuery)
	if err != nil {
		log.Printf("GuessDistrictForZip(): Error preparing Query: %v", err)
		return "", err
	}

	row := statement.QueryRow(zip)
	var district string
	err = row.Scan(&district)
	if err != nil {
		log.Printf("GuessDistrictForZip(): Error querying AK: %v", err)
	}

	return district, err
}

var getEmailsQuery = `
select 
	u.email
from core_user u 
inner join core_phone p on u.id = p.user_id
inner join core_subscription s on u.id = s.user_id
where s.list_id = 9 AND p.phone = ?
group by u.email;
`

// GetEmailsForPhoneNumber grabs all emails corresponding to a particular
// phone number. This is used when unsubscribing or subscribing a user
// when they opt out or in via twilio
func (ak *AKDB) GetEmailsForPhoneNumber(phoneNumber string) ([]string, error) {
	statement, err := ak.Session.Prepare(getEmailsQuery)
	if err != nil {
		log.Printf("GetEmailsForPhoneNumber(): Error preparing Query: %v", err)
		return nil, err
	}

	rows, err := statement.Query(phoneNumber)
	if err == nil {
		// micro-optimization
		var emails []string
		for rows.Next() {
			var email string
			err = rows.Scan(&email)
			if err != nil {
				log.Printf("GetEmailsForPhoneNumber(): Error reading row (continuing): %v", err)
			} else {
				emails = append(emails, email)
			}
		}

		return emails, nil
	}

	log.Printf("GetEmailsForPhoneNumber(): Error querying AK: %v", err)
	return nil, err
}

// Close satisfies the AK interface. Closes the DB connection. This should only be
// called when the application resolves.
func (ak *AKDB) Close() {
	ak.Session.Close()
}
