package remote

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/internal/logger"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/stretchr/testify/require"
)

func TestClient_FetchRetryWithDifferentResponseCodes(t *testing.T) {
	// Test data: Response code, error message, and expected number of fetch calls
	testData := []struct {
		responseCode int
		errorMessage string
		fetchCalls   int
	}{
		{300, "Fetch Exception 300", 2},
		{400, "Fetch Exception 400", 1},
		{429, "Fetch Exception 429", 2},
		{500, "Fetch Exception 500", 2},
		{0, "Other Exception", 2},
	}

	for _, data := range testData {
		// Mock client initialization with httptest
		config := &Config{
			FetchTimeout: 500 * time.Millisecond,
			Debug:        true,
			RetryBackoff: &RetryBackoff{
				FetchRetries:      1,
				FetchRetryTimeout: 500 * time.Millisecond,
			},
		}

		// Variable to track the number of requests
		requestCount := 0

		// Create a new httptest.Server for each iteration
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment the request count
			requestCount++
			// Mock the doFetch method to throw FetchException or other exceptions
			if requestCount == 1 {
				// For the first request, return the specified error and status code
				http.Error(w, data.errorMessage, data.responseCode)
			} else if requestCount == 2 {
				// For the second request, return a 200 response
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("{}"))
				if err != nil {
					return
				}
			}
		}))

		// Update the client config to use the test server
		config.ServerUrl = server.URL
		client := &Client{
			log:    logger.New(config.Debug),
			apiKey: "apiKey",
			config: config,
			client: server.Client(), // Use the test server's client
		}

		fmt.Printf("%d %s\n", data.responseCode, data.errorMessage)

		// Perform the fetch and catch the exception
		_, err := client.Fetch(&experiment.User{UserId: "test_user"})
		if err != nil {
			// catch exception
		}

		// Close the server
		server.Close()

		// Assert the expected number of requests
		require.Equal(t, data.fetchCalls, requestCount, "Unexpected number of requests")
	}
}
