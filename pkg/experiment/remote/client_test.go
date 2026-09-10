package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/amplitude/experiment-go-server/pkg/logger"
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

func TestClient_FetchV2WithOptions_FlagKeys_Partial(t *testing.T) {
	client := Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz", nil)
	user := &experiment.User{UserId: "test_user"}
	result, err := client.FetchV2WithOptions(user, &FetchOptions{
		FlagKeys:         []string{"sdk-ci-test"},
		TracksAssignment: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	variant := result["sdk-ci-test"]
	require.Equal(t, "on", variant.Key)
	require.Equal(t, "on", variant.Value)
	require.Equal(t, "payload", variant.Payload)
}

func TestClient_FetchV2WithOptions_FlagKeys_None_ReturnsAll(t *testing.T) {
	client := Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz", nil)
	user := &experiment.User{UserId: "test_user"}
	result, err := client.FetchV2WithOptions(user, &FetchOptions{
		TracksAssignment: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result), 2)
}

func TestClient_FetchV2WithOptions_FlagKeys_NonExistent(t *testing.T) {
	client := Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz", nil)
	user := &experiment.User{UserId: "test_user"}
	result, err := client.FetchV2WithOptions(user, &FetchOptions{
		FlagKeys:         []string{"123"},
		TracksAssignment: true,
	})
	require.NoError(t, err)
	require.Empty(t, result)
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

func TestClient_UserSuppliedConfig_FetchLoggingDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Fetch Exception 500", http.StatusInternalServerError)
	}))
	defer server.Close()

	testData := []struct {
		name   string
		config *Config
	}{
		// Debug logs are emitted before the request is even sent.
		{"debug enabled", &Config{Debug: true}},
		// Nothing is logged until the fetch fails at the default error level.
		{"default log level", &Config{}},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			data.config.ServerUrl = server.URL
			config := fillConfigDefaults(data.config)
			client := &Client{
				log:    logger.New(config.LogLevel, config.LoggerProvider),
				apiKey: "apiKey",
				config: config,
				client: server.Client(),
			}

			require.NotPanics(t, func() {
				_, _ = client.Fetch(&experiment.User{UserId: "test_user"})
			})
		})
	}
}

func TestInitialize_UserSuppliedConfig_DoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		Initialize("apiKey-user-supplied-config", &Config{Debug: true})
	})
}

// recordingLoggerProvider records debug messages and delegates the rest to the default provider.
type recordingLoggerProvider struct {
	logger.LoggerProvider
	debugMessages []string
}

func (p *recordingLoggerProvider) Debug(message string, args ...interface{}) {
	p.debugMessages = append(p.debugMessages, message)
}

func TestInitialize_UserSuppliedLoggerProvider_IsUsed(t *testing.T) {
	provider := &recordingLoggerProvider{LoggerProvider: logger.NewDefault()}

	Initialize("apiKey-user-supplied-logger-provider", &Config{Debug: true, LoggerProvider: provider})

	require.NotEmpty(t, provider.debugMessages, "expected the supplied logger provider to receive the debug logs")
}

func TestClient_FetchV2WithOptions(t *testing.T) {
	testData := []FetchOptions{
		{TracksAssignment: true, TracksExposure: true},
		{FlagKeys: []string{}, TracksAssignment: true, TracksExposure: false},
		{FlagKeys: []string{"flag-1"}, TracksAssignment: false, TracksExposure: true},
		{FlagKeys: []string{"flag-1", "flag-2"}, TracksAssignment: false, TracksExposure: false},
	}

	for _, fetchOptions := range testData {
		// Create a new httptest.Server for each iteration
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if fetchOptions.FlagKeys == nil {
				require.Empty(t, r.Header.Get("X-Amp-Exp-Flag-Keys"))
			} else {
				flagKeysJSON, err := json.Marshal(fetchOptions.FlagKeys)
				require.NoError(t, err)
				require.Equal(
					t,
					base64.RawURLEncoding.EncodeToString(flagKeysJSON),
					r.Header.Get("X-Amp-Exp-Flag-Keys"),
				)
			}
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
			log:    logger.New(logger.Debug, logger.NewDefault()),
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
