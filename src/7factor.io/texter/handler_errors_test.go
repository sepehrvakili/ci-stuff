package texter

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Internal tests for un-exported methods.
func Test_RenderingErrorsWorks(context *testing.T) {
	expectedMessage := `{ "status": 400, "message": "This is a test" }`
	recorder := httptest.NewRecorder()
	renderErrorWithHTTPStatus(recorder, errors.New("This is a test"), 400)

	if recorder.Body.String() != expectedMessage {
		context.Errorf("Response doesn't match expected. got %s want %s",
			recorder.Body.String(), expectedMessage)
	}
}

func Test_HTMLEncodingErrorsWorks(context *testing.T) {
	expectedMessage := `{ "status": 400, "message": "\u003chtml\u003eIn this junk\u003c/html\u003e" }`
	recorder := httptest.NewRecorder()
	renderErrorWithHTTPStatus(recorder, errors.New("<html>In this junk</html>"), 400)

	if recorder.Body.String() != expectedMessage {
		context.Errorf("Response doesn't match expected. got %s want %s",
			recorder.Body.String(), expectedMessage)
	}
}

func Test_RenderTwilioException(context *testing.T) {
	expectedStatus := http.StatusBadRequest
	err := TwilioException{
		Status:   expectedStatus,
		Message:  "Bad juju",
		Code:     9999,
		MoreInfo: "Badder juju",
	}

	buff, _ := json.Marshal(&err)
	expectedExceptionJSON := string(buff)

	recorder := httptest.NewRecorder()
	renderError(recorder, err)

	if recorder.Code != expectedStatus {
		context.Errorf("Response doesn't match expected. got %v want %v",
			recorder.Code, expectedStatus)
	}

	// Curiously note that the recorder always has a trailing space on it. So, in order
	// to assert equality you need to trim your body before comparing it.
	if body := strings.TrimSpace(recorder.Body.String()); strings.Compare(body, expectedExceptionJSON) != 0 {
		context.Errorf("Body doesn't match expected. got %v want %v",
			body, expectedExceptionJSON)
	}
}

func Test_RenderingUnknownError(context *testing.T) {
	err := errors.New("random error")
	recorder := httptest.NewRecorder()
	renderError(recorder, err)

	if recorder.Code != http.StatusInternalServerError {
		context.Errorf("Response doesn't match expected. got %v want %v",
			recorder.Code, http.StatusInternalServerError)
	}
}
