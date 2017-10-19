package texter

import (
	"database/sql"
	"log"
	"time"

	// for the sql drivers
	_ "github.com/go-sql-driver/mysql"
)

// Congress is an interface to action kit data.
type Congress interface {
	GetRepresentatives() (Representatives, error)
	Close()
}

// CongressDB manages connections and queries to ActionKit's MySQL for our
// application.
type CongressDB struct {
	Creds   DBCreds
	Session *sql.DB
}

// NewCongressDB returns a new ActionKit interface object.
func NewCongressDB(creds DBCreds) (Congress, error) {
	log.Printf("connecting to CongressDB at %v", creds.Host)
	db, _ := sql.Open("mysql", creds.ToDSN())
	err := db.Ping() // mCongressDBe sure the DB works

	if err != nil {
		// Let's wait for a real connection, 20 retries
		for count := 1; count <= 20; count++ {
			log.Printf("retrying connection...")
			time.Sleep(5 * time.Second)
			err = db.Ping()

			if err == nil {
				log.Printf("connected to DB at URL %v", creds.Host)
				return &CongressDB{Creds: creds, Session: db}, err
			}
		}

		log.Printf("unable to connect to the CongressDB DB, giving up!")
		return nil, err
	}

	log.Printf("connected to DB at URL %v", creds.Host)
	return &CongressDB{Creds: creds, Session: db}, err
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

// GetRepresentatives satisfies the CongressDB interface. Returns a Representative object
// that is a map of district to representative.
func (db *CongressDB) GetRepresentatives() (Representatives, error) {
	reps := make(Representatives)
	rows, err := db.Session.Query(repsQuery)
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
		log.Printf("GetRepresentatives(): Error querying CongressDB: %v", err)
	}
	return reps, err
}

// Close satisfies the CongressDB interface. Closes the DB connection. This should only be
// called when the application resolves.
func (db *CongressDB) Close() {
	db.Session.Close()
}
