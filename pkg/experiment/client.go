package experiment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var client *Client
var mutex = sync.Once{}

type Client struct {
	apiKey string
	config *Config
	client *http.Client
}

func Initialize(apiKey string, config *Config) *Client {
	mutex.Do(func() {
		if apiKey == "" {
			panic("api key must be set")
		}
		config = fillConfigDefaults(config)
		client = &Client{
			apiKey: apiKey,
			config: config,
			client: &http.Client{},
		}
	})
	return client
}

func (c *Client) Fetch(user *User) (Variants, error) {
	variants, err := c.doFetch(user, c.config.FetchTimeoutMillis)
	if err != nil {
		return c.retryFetch(user)
	}
	return variants, err
}

func (c *Client) doFetch(user *User, timeoutMillis int) (Variants, error) {
	addLibraryContext(user)
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/vardata"
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMillis) * time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch resulted in error response %v", resp.StatusCode)
	}
	return parseResponse(resp)
}

func (c *Client) retryFetch(user *User) (Variants, error) {
	var err error
	var timer *time.Timer
	delay := time.Duration(c.config.RetryBackoff.FetchRetryBackoffMinMillis) * time.Millisecond
	for i := 0; i < c.config.RetryBackoff.FetchRetries; i++ {
		timer = time.NewTimer(delay)
		<- timer.C
		variants, err := c.doFetch(user, c.config.RetryBackoff.FetchRetryTimeoutMillis)
		if err == nil && variants != nil {
			return variants, nil
		}
	}
	return nil, err
}

func addLibraryContext(user *User) {
	if user.Library == "" {
		user.Library = fmt.Sprintf("experiment-go-server/%v", VERSION)
	}
}

func parseResponse(resp *http.Response) (Variants, error) {
	interop := make(interopVariants)
	err := json.NewDecoder(resp.Body).Decode(&interop)
	if err != nil {
		return nil, err
	}
	variants := make(Variants)
	for k, iv := range interop {
		var value string
		if iv.Value != "" {
			value = iv.Value
		} else if iv.Key != "" {
			value = iv.Value
		}
		variants[k] = Variant{
			Value: value,
			Payload: iv.Payload,
		}
	}
	return variants, nil
}