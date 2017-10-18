package texter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetHealthCheckShouldReturnValidJSON(context *testing.T) {
	theTest := TestParameters{
		http.MethodGet,
		"/status",
		"/status",
		nil,
		HealthCheck(),
		func(context *testing.T, recorder *httptest.ResponseRecorder, request *http.Request) {
			if status := recorder.Code; status != http.StatusOK {
				context.Errorf("Handler returned an incorrect status code: got %v want %v",
					status, http.StatusOK)
			}

			expected := `{"gtg":true}`
			if recorder.Body.String() != expected {
				context.Errorf("Handler returned unexpected body: got %v want %v",
					recorder.Body.String(), expected)
			}
		},
	}
	Run(context, theTest)
}
