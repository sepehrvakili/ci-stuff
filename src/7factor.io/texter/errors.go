package texter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ErrorCollector is a convenience type
type ErrorCollector []error

// Collect grabs errors for storage
func (c *ErrorCollector) Collect(e error) {
	*c = append(*c, e)
}

// Error satisfies the error interface
func (c *ErrorCollector) Error() (err string) {
	err = "Collected errors:\n"
	for i, e := range *c {
		err += fmt.Sprintf("\tError %d: %s\n", i, e.Error())
	}

	return err
}

// Private, package level functions for handling error messages being
// written in JSON to a writer buffer. Simple. I like this better than
// hacking some interface weirdness.
var errorTemplate = `{ "status": %v, "message": "%s" }`

func renderErrorWithHTTPStatus(writer http.ResponseWriter, err error, status int) {
	writer.WriteHeader(status)
	buffer := new(bytes.Buffer)
	json.HTMLEscape(buffer, []byte(err.Error()))
	log.Printf(errorTemplate, status, buffer)
	fmt.Fprintf(writer, errorTemplate, status, buffer)
}

// Private function to translate the return codes coming from other sources.
// It translates errors to appropriate HTTP statuses.
func renderError(writer http.ResponseWriter, err error) {
	switch err.(type) {
	case TwilioException:
		twilioErr := err.(TwilioException)
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoderErr := encoder.Encode(&twilioErr)

		if encoderErr != nil {
			renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
		} else {
			writer.WriteHeader(twilioErr.Status)
			writer.Write(buffer.Bytes())
			log.Print(buffer.String())
		}
	default:
		renderErrorWithHTTPStatus(writer, err, http.StatusInternalServerError)
	}
}
