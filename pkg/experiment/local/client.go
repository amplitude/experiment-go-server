package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amplitude/analytics-go/amplitude"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/amplitude/experiment-go-server/internal/evaluation"

	"github.com/amplitude/experiment-go-server/pkg/experiment"

	"github.com/amplitude/experiment-go-server/internal/logger"
)

var clients = map[string]*Client{}
var initMutex = sync.Mutex{}

type Client struct {
	log               *logger.Log
	apiKey            string
	config            *Config
	client            *http.Client
	poller            *poller
	flags             map[string]interface{}
	assignmentService *assignmentService
}

func Initialize(apiKey string, config *Config) *Client {
	initMutex.Lock()
	client := clients[apiKey]
	if client == nil {
		if apiKey == "" {
			panic("api key must be set")
		}
		config = fillConfigDefaults(config)
		client = &Client{
			log:    logger.New(config.Debug),
			apiKey: apiKey,
			config: config,
			client: &http.Client{},
			poller: newPoller(),
		}
		client.log.Debug("config: %v", *config)
	}
	// create assignment service if apikey is provided
	if config.AssignmentConfig != nil && config.AssignmentConfig.IsValid() {
		instance := amplitude.NewClient(config.AssignmentConfig.Config)
		filter := newAssignmentFilter(config.AssignmentConfig.CacheCapacity)
		client.assignmentService = &assignmentService{
			amplitude: &instance, filter: filter,
		}
	}
	initMutex.Unlock()
	return client
}

func (c *Client) Start() error {
	result, err := c.doFlags()
	if err != nil {
		return err
	}
	c.flags = result
	c.poller.Poll(c.config.FlagConfigPollerInterval, func() {
		result, err := c.doFlags()
		if err != nil {
			return
		}
		c.flags = result
	})
	return nil
}

func (c *Client) Evaluate(user *experiment.User, flagKeys []string) (map[string]experiment.Variant, error) {
	variants := make(map[string]experiment.Variant)
	if len(c.flags) == 0 {
		c.log.Debug("evaluate: no flags")
		return variants, nil
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	sortedFlags, err := topologicalSort(c.flags, flagKeys)
	if err != nil {
		return nil, err
	}
	flagsJson, err := json.Marshal(sortedFlags)
	if err != nil {
		return nil, err
	}
	c.log.Debug("evaluate:\n\t- user: %v\n\t- rules: %v\n", string(userJson), string(flagsJson))
	resultJson := evaluation.Evaluate(string(flagsJson), string(userJson))
	c.log.Debug("evaluate result: %v\n", resultJson)
	var interopResult *interopResult
	err = json.Unmarshal([]byte(resultJson), &interopResult)
	if err != nil {
		return nil, err
	}
	if interopResult.Error != nil {
		return nil, fmt.Errorf("evaluation resulted in error: %v", *interopResult.Error)
	}
	result := interopResult.Result
	assignmentResult := evaluationResult{}

	filter := len(flagKeys) != 0
	for k, v := range *result {
		included := !filter || contains(flagKeys, k)
		if !v.IsDefaultVariant && included {
			variants[k] = experiment.Variant{
				Value:   v.Variant.Key,
				Payload: v.Variant.Payload,
			}
		}
		if included || v.Type == flagTypeMutualExclusionGroup || v.Type == flagTypeHoldoutGroup {
			assignmentResult[k] = v
		}
	}
	if c.assignmentService != nil {
		(*c.assignmentService).Track(newAssignment(user, &assignmentResult))
	}
	return variants, nil
}

func (c *Client) Rules() (map[string]interface{}, error) {
	return c.doRules()
}

func (c *Client) doRules() (map[string]interface{}, error) {
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/rules"
	endpoint.RawQuery = "eval_mode=local"
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FlagConfigPollerRequestTimeout)
	defer cancel()
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.log.Debug("rules: %v", string(body))
	var rules []map[string]interface{}
	err = json.Unmarshal(body, &rules)
	if err != nil {
		return nil, err
	}
	var result = make(map[string]interface{})
	for _, rule := range rules {
		flagKey := rule["flagKey"]
		result[fmt.Sprintf("%v", flagKey)] = rule
	}
	return result, nil
}

func (c *Client) Flags() (*string, error) {
	flags, err := c.doFlags()
	if err != nil {
		return nil, err
	}
	flagsJson, err := json.Marshal(flags)
	if err != nil {
		return nil, err
	}
	flagsString := string(flagsJson)
	return &flagsString, nil
}

func (c *Client) doFlags() (map[string]interface{}, error) {
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/v1/flags"
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FlagConfigPollerRequestTimeout)
	defer cancel()
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.log.Debug("flags: %v", string(body))
	flagsArray := make([]interface{}, 0)
	err = json.Unmarshal(body, &flagsArray)
	if err != nil {
		return nil, err
	}
	// Extract keys and create flags map
	flags := make(map[string]interface{})
	for _, flagAny := range flagsArray {
		switch flag := flagAny.(type) {
		case map[string]interface{}:
			switch flagKey := flag["flagKey"].(type) {
			case string:
				flags[flagKey] = flag
			}
		}
	}
	return flags, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
