package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/amplitude/experiment-go-server/internal/evaluation"

	"github.com/amplitude/experiment-go-server/pkg/experiment"

	"github.com/amplitude/experiment-go-server/internal/logger"
)

var client *Client
var once = sync.Once{}
var evaluationMutex = sync.Mutex{}

type Client struct {
	log    *logger.Log
	apiKey string
	config *Config
	client *http.Client
	poller *poller
	rules  map[string]interface{}
	mutex  *sync.Mutex
}

func Initialize(apiKey string, config *Config) *Client {
	once.Do(func() {
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
	})
	return client
}

func (c *Client) Start() error {
	result, err := c.doRules()
	if err != nil {
		return err
	}
	c.rules = result
	c.poller.Poll(c.config.FlagConfigPollerInterval, func() {
		result, err := c.doRules()
		if err != nil {
			return
		}
		c.rules = result
	})

	return nil
}

func (c *Client) Evaluate(user *experiment.User, flagKeys []string) (map[string]experiment.Variant, error) {
	noFlagKeys := flagKeys == nil || len(flagKeys) == 0
	rules := make([]interface{}, 0)
	for k, v := range c.rules {
		if noFlagKeys || contains(flagKeys, k) {
			rules = append(rules, v)
		}
	}
	rulesJson, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	c.log.Debug("evaluate:\n\t- user: %v\n\t- rules: %v\n", string(userJson), string(rulesJson))

	evaluationMutex.Lock()
	resultJson := evaluation.Evaluate(string(rulesJson), string(userJson))
	evaluationMutex.Unlock()
	var result *evaluationResult
	err = json.Unmarshal([]byte(resultJson), &result)
	if err != nil {
		return nil, err
	}
	variants := make(map[string]experiment.Variant)
	for k, v := range *result {
		variants[k] = experiment.Variant{
			Value:   v.Variant.Key,
			Payload: v.Variant.Payload,
		}
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
