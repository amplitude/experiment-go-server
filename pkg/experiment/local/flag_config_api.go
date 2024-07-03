package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type flagConfigApi interface {
	GetFlagConfigs() (map[string]*evaluation.Flag, error)
}

type flagConfigApiV2 struct {
	DeploymentKey                        string
	ServerURL                            string
	FlagConfigPollerRequestTimeoutMillis time.Duration
}

func newFlagConfigApiV2(deploymentKey, serverURL string, flagConfigPollerRequestTimeoutMillis time.Duration) *flagConfigApiV2 {
	return &flagConfigApiV2{
		DeploymentKey:                        deploymentKey,
		ServerURL:                            serverURL,
		FlagConfigPollerRequestTimeoutMillis: flagConfigPollerRequestTimeoutMillis,
	}
}

func (a *flagConfigApiV2) GetFlagConfigs() (map[string]*evaluation.Flag, error) {
	client := &http.Client{}
	endpoint, err := url.Parse("https://api.lab.amplitude.com/")
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/v2/flags"
	endpoint.RawQuery = "v=0"
	ctx, cancel := context.WithTimeout(context.Background(), a.FlagConfigPollerRequestTimeoutMillis)
	defer cancel()
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", a.DeploymentKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var flagsArray []*evaluation.Flag
	err = json.Unmarshal(body, &flagsArray)
	if err != nil {
		return nil, err
	}
	flags := make(map[string]*evaluation.Flag)
	for _, flag := range flagsArray {
		flags[flag.Key] = flag
	}
	return flags, nil
}
