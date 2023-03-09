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

type Client struct {
	log    *logger.Log
	apiKey string
	config *Config
	client *http.Client
	poller *poller
	flags  *string
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
	if len(*c.flags) == 0 {
		c.log.Debug("evaluate: no flags")
		return variants, nil

	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	c.log.Debug("evaluate:\n\t- user: %v\n\t- rules: %v\n", string(userJson), *c.flags)

	resultJson := evaluation.Evaluate(*c.flags, string(userJson))
	c.log.Debug("evaluate result: %v\n", resultJson)
	var result *evaluationResult
	err = json.Unmarshal([]byte(resultJson), &result)
	if err != nil {
		return nil, err
	}
	filter := flagKeys != nil && len(flagKeys) != 0
	for k, v := range *result {
		if v.IsDefaultVariant || (filter && !contains(flagKeys, k)) {
			continue
		}
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
	return c.doFlags()
}

func (c *Client) doFlags() (*string, error) {
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
	flags := string(body)
	c.log.Debug("flags: %v", flags)
	return &flags, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
