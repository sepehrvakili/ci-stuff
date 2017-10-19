package texter

// Target is someone who is going to receive a text.
type Target struct {
	PhoneNumber string
	USDistrict  string
	Body        string
}

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
