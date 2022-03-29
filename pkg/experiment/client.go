package experiment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/internal/logger"
	"github.com/gorilla/websocket"
)

var client *Client
var once = sync.Once{}

type Client struct {
	log    *logger.Log
	apiKey string
	config *Config
	client *http.Client
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
		}
		client.log.Debug("config: %v", *config)
	})
	return client
}

func (c *Client) Fetch(user *User) (Variants, error) {
	variants, err := c.doFetch(user, c.config.FetchTimeout)
	if err != nil {
		c.log.Error("fetch error: %v", err)
		if c.config.RetryBackoff.FetchRetries > 0 {
			return c.retryFetch(user)
		} else {
			return nil, err
		}
	}
	return variants, err
}

func (c *Client) Rules() (string, error) {
	return c.doRules()
}

func (c *Client) Stream(user *User) (chan Variants, error) {
	return c.doStream(user, c.config.FetchTimeout)
}

func (c *Client) Publish(key, value string) error {
	return c.doPublish(key, value)
}

func (c *Client) doFetch(user *User, timeout time.Duration) (Variants, error) {
	addLibraryContext(user)
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/vardata"
	endpoint.RawQuery = fmt.Sprintf("d=%s", randStringRunes(5))
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	c.log.Debug("fetch variants for user %s", string(jsonBytes))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("fetch request resulted in error response %v", resp.StatusCode)
	}
	return c.parseResponse(resp)
}

func (c *Client) doRules() (string, error) {
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return "", err
	}
	endpoint.Path = "sdk/rules"
	endpoint.RawQuery = "eval_mode=local"
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) doStream(user *User, timeout time.Duration) (chan Variants, error) {
	addLibraryContext(user)
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Scheme = "ws"
	endpoint.Path = "stream/subscribe"
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	c.log.Debug("fetch variants for user %s", string(jsonBytes))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	headers := http.Header{}
	headers.Set("X-Amp-Exp-User", base64.StdEncoding.EncodeToString(jsonBytes))
	headers.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, endpoint.String(), headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("fetch request resulted in error response %v", resp.StatusCode)
	}
	c.log.Debug("fetch success: %v", *resp)
	variants := make(chan Variants)
	go func() {
		for {
			msg := make(interopVariants)
			err := conn.ReadJSON(&msg)
			if err != nil {
				close(variants)
				return
			}
			variants <- c.convertInteropVariants(msg)
		}
	}()
	return variants, nil
}

func (c *Client) doPublish(key, value string) error {
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return err
	}
	endpoint.Path = "stream/publish"
	reqBody := &publishRequest{
		Key:   key,
		Value: value,
	}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	c.log.Debug("publish %s", string(jsonBytes))
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("publish request resulted in error response %v", resp.StatusCode)
	}
	c.log.Debug("publish success: %v", *resp)
	return nil
}

func (c *Client) retryFetch(user *User) (Variants, error) {
	var err error
	var timer *time.Timer
	delay := c.config.RetryBackoff.FetchRetryBackoffMin
	for i := 0; i < c.config.RetryBackoff.FetchRetries; i++ {
		c.log.Debug("retry attempt %v", i)
		timer = time.NewTimer(delay)
		<-timer.C
		variants, err := c.doFetch(user, c.config.RetryBackoff.FetchRetryTimeout)
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

func (c *Client) parseResponse(resp *http.Response) (Variants, error) {
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
			value = iv.Key
		}
		variants[k] = Variant{
			Value:   value,
			Payload: iv.Payload,
		}
	}
	c.log.Debug("parsed variants from response: %v", variants)
	return variants, nil
}

func (c *Client) convertInteropVariants(interop interopVariants) Variants {
	variants := make(Variants)
	for k, iv := range interop {
		var value string
		if iv.Value != "" {
			value = iv.Value
		} else if iv.Key != "" {
			value = iv.Key
		}
		variants[k] = Variant{
			Value:   value,
			Payload: iv.Payload,
		}
	}
	c.log.Debug("parsed variants from response: %v", variants)
	return variants
}

func addLibraryContext(user *User) {
	if user.Library == "" {
		user.Library = fmt.Sprintf("experiment-go-server/%v", VERSION)
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
