package texter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Message represents anything you want to send to twilio.
type Message struct {
	To   string
	From string
	Body string
}

// URLEncode transforms a message to something Twilio understands
func (m Message) URLEncode() io.Reader {
	values := url.Values{}
	values.Set("To", m.To)
	values.Set("From", m.From)
	values.Set("Body", m.Body)

	mss := os.Getenv("TW_MSGSID")
	if mss != "" {
		values.Set("MessagingServiceSid", os.Getenv("TW_MSGSID"))
	}

	log.Printf("values: %v", values)
	return strings.NewReader(values.Encode())
}

// Segments will create a slice of messages from a single message.
// This exists as a work around for SMS messages that are not
// concatenated inside carrier networks (maily only due to sprint)
func (m Message) Segments() []Message {
	chunker := NewMessageChunker()
	splits := chunker.Split(m.Body)

	if len(splits) > 1 {
		var messages []Message
		for _, item := range splits {
			messages = append(messages, Message{To: m.To, From: m.From, Body: item})
		}
		return messages
	}

	return []Message{m}
}

// HTTPClient is an interface for mocking HTTP clients. This is satisfied by the GoLang
// internal client and our mocks.
type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// TwilioClient represents a set of methods that can send information to
// twilio.
type TwilioClient interface {
	SendSMS(msg *Message) (int, error)
}

// Twilio represents information you need to know about in order to
// communicate to the service.
type Twilio struct {
	AccountSid string
	AuthToken  string
	BaseURL    string
	Client     HTTPClient
}

// TwilioException is a representation of a twilio exception.
type TwilioException struct {
	Status   int    `json:"status"`    // HTTP specific error code
	Message  string `json:"message"`   // HTTP error message
	Code     int    `json:"code"`      // Twilio specific error code
	MoreInfo string `json:"more_info"` // Additional info from Twilio
}

// TwilioMessage represents an incoming message from Twilio.
type TwilioMessage struct {
	MessageSid string
	AccountSid string
	From       string
	To         string
	Body       string
}

// NewTwilioMessage creates a message struct from the values parsed
// out of a query string.
func NewTwilioMessage(values url.Values) TwilioMessage {
	return TwilioMessage{
		MessageSid: values.Get("MessageSid"),
		AccountSid: values.Get("AccountSid"),
		From:       CleanPhoneNumber(values.Get("From")),
		To:         CleanPhoneNumber(values.Get("To")),
		Body:       values.Get("Body"),
	}
}

// Error makes Exception conform to the standard GoLang Error interface.
func (e TwilioException) Error() string {
	return e.Message
}

// NewTW builds out a new Twilio client from the environment
func NewTW() Twilio {
	return Twilio{
		AccountSid: os.Getenv("TW_ACCT"),
		AuthToken:  os.Getenv("TW_TOKEN"),
		BaseURL:    fmt.Sprintf(os.Getenv("TW_URL"), os.Getenv("TW_ACCT")),
		Client:     new(http.Client),
	}
}

// SendSMS satisfies the TwilioClient interface for the Twilio object. It will
// send an SMS through Twilio to the target user with the appropriate
// message.
func (c Twilio) SendSMS(msg *Message) (int, error) {
	throttle, err := strconv.Atoi(os.Getenv("SEGMENT_TICKINTERVAL"))
	if err != nil {
		throttle = 1000
		log.Printf("Using default value for throttle: %v", throttle)
	}

	// Very simple regex
	regex, err := regexp.Compile(`[1]?[0-9]{10}`)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	if !regex.MatchString(msg.To) || !regex.MatchString(msg.From) {
		errorMsg := fmt.Sprintf("got a bad phone number: to [%v] from [%v]", msg.To, msg.From)
		return http.StatusBadRequest, errors.New(errorMsg)
	}

	segments := msg.Segments()

	var code int
	for _, segment := range segments {
		code, err = c.sendRequest(&segment)

		if err != nil {
			return code, err
		}

		time.Sleep(time.Duration(throttle) * time.Millisecond)
	}

	return code, err
}

func (c Twilio) sendRequest(msg *Message) (int, error) {

	if c.Client == nil {
		return http.StatusInternalServerError, errors.New("http client is nil, cannot proceed")
	}

	request, err := http.NewRequest(http.MethodPost, c.BaseURL, msg.URLEncode())
	request.SetBasicAuth(c.AccountSid, c.AuthToken)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	log.Printf("Sending message to twilio: %#v", msg)

	if err == nil {
		response, err := c.Client.Do(request)

		// in this case, a problem occured during the request.
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return handleTwilioResponse(response)
	}

	// Something went wrong, client is nil. Falls through to this.
	return http.StatusInternalServerError, err
}

// CleanPhoneNumber is a utility function for stripping a phone number of useless stuff.
func CleanPhoneNumber(number string) string {
	toReplace := "()- +"
	for _, r := range toReplace {
		number = strings.Replace(number, string(r), "", -1)
	}

	// This number includes the "1"
	if len(number) == 11 && strings.HasPrefix(number, "1") {
		return number[1:len(number)]
	}

	return number
}

// Private function for transposing twilio responses.
func handleTwilioResponse(response *http.Response) (int, error) {
	switch response.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusCreated:
		return response.StatusCode, nil
	case http.StatusBadRequest:
		var exc TwilioException
		err := json.NewDecoder(response.Body).Decode(&exc)
		if err != nil {
			bytes, _ := ioutil.ReadAll(response.Body)
			log.Printf("unable to handle message %v decoder error %v", string(bytes), err)
			return http.StatusInternalServerError, err
		}
		return exc.Status, exc
	case http.StatusUnauthorized:
		return http.StatusUnauthorized, errors.New("unauthorized")
	case http.StatusNotFound:
		return http.StatusNotFound, errors.New("couldn't find what you were looking for")
	case http.StatusMethodNotAllowed:
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	case http.StatusTooManyRequests:
		return http.StatusTooManyRequests, errors.New("too many requests")
	}

	return http.StatusInternalServerError, errors.New("unknown twilio response")
}
