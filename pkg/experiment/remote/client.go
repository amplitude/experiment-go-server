package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"

	"github.com/amplitude/experiment-go-server/internal/logger"
)

var clients = map[string]*Client{}
var initMutex = sync.Mutex{}

type Client struct {
	log    *logger.Log
	apiKey string
	config *Config
	client *http.Client
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
		}
		client.log.Debug("config: %v", *config)
	}
	initMutex.Unlock()
	return client
}

func (c *Client) Fetch(user *experiment.User) (map[string]experiment.Variant, error) {
	variants, err := c.doFetch(user, c.config.FetchTimeout)
	if err != nil {
		c.log.Error("fetch error: %v", err)
		if c.config.RetryBackoff.FetchRetries > 0 && shouldRetryFetch(err) {
			return c.retryFetch(user)
		} else {
			return nil, err
		}
	}
	return variants, err
}

func (c *Client) doFetch(user *experiment.User, timeout time.Duration) (map[string]experiment.Variant, error) {
	addLibraryContext(user)
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/vardata"
	if c.config.Debug {
		endpoint.RawQuery = fmt.Sprintf("d=%s", randStringRunes(5))
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	c.log.Debug("fetch variants for user %s", string(jsonBytes))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	c.log.Debug("fetch request: %v", req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.log.Debug("fetch response: %v", *resp)
	if resp.StatusCode != http.StatusOK {
		return nil, &experiment.FetchError{StatusCode: resp.StatusCode, Message: resp.Status}
	}
	return c.parseResponse(resp)
}

func (c *Client) retryFetch(user *experiment.User) (map[string]experiment.Variant, error) {
	var err error
	var variants map[string]experiment.Variant
	var timer *time.Timer
	delay := c.config.RetryBackoff.FetchRetryBackoffMin
	for i := 0; i < c.config.RetryBackoff.FetchRetries; i++ {
		c.log.Debug("retry attempt %v", i)
		timer = time.NewTimer(delay)
		<-timer.C
		variants, err = c.doFetch(user, c.config.RetryBackoff.FetchRetryTimeout)
		if err == nil && variants != nil {
			c.log.Debug("retry attempt %v success", i)
			return variants, nil
		}
		c.log.Debug("retry attempt %v error: %v", i, err)
		delay = time.Duration(math.Min(
			float64(delay)*c.config.RetryBackoff.FetchRetryBackoffScalar,
			float64(c.config.RetryBackoff.FetchRetryBackoffMax)),
		)
	}
	c.log.Error("fetch retries failed after %v attempts: %v", c.config.RetryBackoff.FetchRetries, err)
	return nil, err
}

func (c *Client) parseResponse(resp *http.Response) (map[string]experiment.Variant, error) {
	interop := make(interopVariants)
	err := json.NewDecoder(resp.Body).Decode(&interop)
	if err != nil {
		return nil, err
	}
	variants := make(map[string]experiment.Variant)
	for k, iv := range interop {
		var value string
		if iv.Value != "" {
			value = iv.Value
		} else if iv.Key != "" {
			value = iv.Key
		}
		variants[k] = experiment.Variant{
			Value:   value,
			Payload: iv.Payload,
		}
	}
	c.log.Debug("parsed variants from response: %v", variants)
	return variants, nil
}

func addLibraryContext(user *experiment.User) {
	if user.Library == "" {
		user.Library = fmt.Sprintf("experiment-go-server/%v", experiment.VERSION)
	}
}

// Helper

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func shouldRetryFetch(err error) bool {
	if err, ok := err.(*experiment.FetchError); ok {
		return err.StatusCode < 400 || err.StatusCode >= 500 || err.StatusCode == 429
	}
	return true
}
