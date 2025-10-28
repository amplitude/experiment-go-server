package remote

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/logger"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/stretchr/testify/require"
)

func TestClient_Fetch_DoesNotReturnDefaultVariants(t *testing.T) {
	client := Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz", nil)
	user := &experiment.User{}
	result, err := client.Fetch(user)
	require.NoError(t, err)
	require.NotNil(t, result)
	variant := result["sdk-ci-test"]
	require.Empty(t, variant)
}

func TestClient_FetchV2_ReturnsDefaultVariants(t *testing.T) {
	client := Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz", nil)
	user := &experiment.User{}
	result, err := client.FetchV2(user)
	require.NoError(t, err)
	require.NotNil(t, result)
	variant := result["sdk-ci-test"]
	require.NotNil(t, variant)
	require.Equal(t, "off", variant.Key)
}

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
			} else {
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
			log:    logger.New(logger.Debug, logger.NewDefault()),
			apiKey: "apiKey",
			config: config,
			client: server.Client(), // Use the test server's client
		}

		fmt.Printf("%d %s\n", data.responseCode, data.errorMessage)

		// Perform the fetch and catch the exception
		_, err := client.Fetch(&experiment.User{UserId: "test_user"})
		if err != nil {
			fmt.Println(err.Error())
		}

		// Close the server
		server.Close()

		// Assert the expected number of requests
		require.Equal(t, data.fetchCalls, requestCount, "Unexpected number of requests")
	}
}

func TestClient_FetchV2WithOptions(t *testing.T) {
	testData := []FetchOptions{
		{TracksAssignment: true, TracksExposure: true},
		{TracksAssignment: true, TracksExposure: false},
		{TracksAssignment: false, TracksExposure: true},
		{TracksAssignment: false, TracksExposure: false},
	}

	for _, fetchOptions := range testData {
		// Create a new httptest.Server for each iteration
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment the request count
			if fetchOptions.TracksAssignment {
				require.Equal(t, r.Header.Get("X-Amp-Exp-Track"), "track")
			} else {
				require.Equal(t, r.Header.Get("X-Amp-Exp-Track"), "no-track")
			}
			if fetchOptions.TracksExposure {
				require.Equal(t, r.Header.Get("X-Amp-Exp-Exposure-Track"), "track")
			} else {
				require.Equal(t, r.Header.Get("X-Amp-Exp-Exposure-Track"), "no-track")
			}
			// Return a 200 response
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				fmt.Println("Response failed")
			}
		}))

		// Update the client config to use the test server
		config := &Config{
			ServerUrl: server.URL,
		}
		fillConfigDefaults(config)
		client := &Client{
			log:    logger.New(config.Debug),
			apiKey: "apiKey",
			config: config,
			client: server.Client(), // Use the test server's client
		}

		_, err := client.FetchV2WithOptions(&experiment.User{UserId: "test_user"}, &fetchOptions)
		if err != nil {
			t.Errorf("FetchV2WithOptions failed: %v", err)
			fmt.Println(err.Error())
		}

		// Close the server
		server.Close()
	}
}
