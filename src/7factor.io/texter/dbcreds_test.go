package texter

import (
	"fmt"
	"os"
	"testing"
)

func Test_DBCredsToDialInfo(context *testing.T) {
	creds := DBCreds{
		Host:   "localhost",
		User:   "jduv",
		Pass:   "abc123",
		DBName: "jduv-db",
	}

	info := creds.ToDialInfo()

	if info.Username != creds.User {
		context.Errorf("Usernames don't match. Got %v expected %v", info.Username, creds.User)
	}

	if info.Password != creds.Pass {
		context.Errorf("Passwords don't match. Got %v expected %v", info.Password, creds.Pass)
	}

	if info.Database != creds.DBName {
		context.Errorf("Databases don't match. Got %v expected %v", info.Database, creds.DBName)
	}

	if len(info.Addrs) != 1 || info.Addrs[0] != creds.Host {
		context.Errorf("DB hosts don't match. Got %v expected %v", info.Addrs, creds.Host)
	}
}

func Test_DBCredsToDialInfoWithMultipleHosts(context *testing.T) {
	creds := DBCreds{
		Host:   "host1,host2,host3",
		User:   "jduv",
		Pass:   "abc123",
		DBName: "jduv-db",
	}

	info := creds.ToDialInfo()

	if len(info.Addrs) != 3 {
		context.Errorf("Number of DBHosts doesn't match. Got %v expected %v", len(info.Addrs), 3)
	}
}

func Test_GetMongoCredentials(context *testing.T) {
	expectedHost := "myhost"
	expectedUser := "jduv"
	expectedPass := "supersecurepw"
	expectedDB := "bubbles"

	os.Setenv("MONGO_URL", expectedHost)
	os.Setenv("MONGO_RRN_USER", expectedUser)
	os.Setenv("MONGO_RRN_PASS", expectedPass)
	os.Setenv("MONGO_RRN_DB", expectedDB)
	creds := GetMongoCredsFromEnv()

	if expectedHost != creds.Host {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedHost, creds.Host)
	}

	if expectedUser != creds.User {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedUser, creds.User)
	}

	if expectedPass != creds.Pass {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedPass, creds.Pass)
	}

	if expectedDB != creds.DBName {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedDB, creds.DBName)
	}
}

func Test_GetAKCredentials(context *testing.T) {
	expectedHost := "myhost"
	expectedUser := "jduv"
	expectedPass := "supersecurepw"
	expectedDB := "bubbles"

	os.Setenv("AK_URL", expectedHost)
	os.Setenv("AK_USER", expectedUser)
	os.Setenv("AK_PASS", expectedPass)
	os.Setenv("AK_DBNAME", expectedDB)
	creds := GetAKCredsFromEnv()

	if expectedHost != creds.Host {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedHost, creds.Host)
	}

	if expectedUser != creds.User {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedUser, creds.User)
	}

	if expectedPass != creds.Pass {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedPass, creds.Pass)
	}

	if expectedDB != creds.DBName {
		context.Errorf("Hostname not set correctly. want %v got %v", expectedDB, creds.DBName)
	}
}

func Test_ToDSN(context *testing.T) {
	expectedHost := "myhost:1234"
	expectedUser := "jduv"
	expectedPass := "supersecurepw"
	expectedDB := "bubbles"
	expectedDSN := fmt.Sprintf("%v:%v@tcp(%v)/%v?timeout=30s", expectedUser, expectedPass,
		expectedHost, expectedDB)

	os.Setenv("AK_URL", expectedHost)
	os.Setenv("AK_USER", expectedUser)
	os.Setenv("AK_PASS", expectedPass)
	os.Setenv("AK_DBNAME", expectedDB)
	creds := GetAKCredsFromEnv()
	dsn := creds.ToDSN()

	if dsn != expectedDSN {
		context.Errorf("DSN's do not match. Got %v wanted %v", dsn, expectedDSN)
	}
}
