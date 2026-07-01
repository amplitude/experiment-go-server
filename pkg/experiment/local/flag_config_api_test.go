package local

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/stretchr/testify/assert"
)

func TestFlagConfigApiV2_GetFlagConfigs_BasePathPreserved(t *testing.T) {
	var receivedPath string
	flags := []*evaluation.Flag{{Key: "flag-1"}}
	flagsJSON, _ := json.Marshal(flags)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write(flagsJSON)
	}))
	defer server.Close()

	testCases := []struct {
		name      string
		serverURL string
	}{
		{"no trailing slash", server.URL + "/p/ff"},
		{"trailing slash", server.URL + "/p/ff/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := newFlagConfigApiV2("deployment-key", tc.serverURL, 5*time.Second)
			result, err := api.getFlagConfigs()
			assert.NoError(t, err)
			assert.Equal(t, "/p/ff/sdk/v2/flags", receivedPath)
			assert.Contains(t, result, "flag-1")
		})
	}
}
