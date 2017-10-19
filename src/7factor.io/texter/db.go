package texter

import (
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
)

// DB interface abstracts the storage mechanism of our application. We
// don't care what kind of DB backs the calls, just so long as we get
// the objects we expect.
type DB interface {
	StowMessage(msg Message) error
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

// StowMessage saves twilio messages to a LostAndFound table
func (db MongoDB) StowMessage(msg Message) error {
	collection := db.Session.DB(db.Creds.DBName).C("Messages")
	return collection.Insert(msg)
}

// Close message shuts down the DB handle.
func (db MongoDB) Close() {
	db.Session.Close()
}
