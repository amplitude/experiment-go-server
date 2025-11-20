package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"

	"github.com/amplitude/experiment-go-server/pkg/logger"
)

var clients = map[string]*Client{}
var initMutex = sync.Mutex{}

type Client struct {
	log    *logger.Logger
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
			log:    logger.New(config.LogLevel, config.LoggerProvider),
			apiKey: apiKey,
			config: config,
			client: &http.Client{},
		}
		client.log.Debug("config: %v", *config)
		clients[apiKey] = client
	}
	initMutex.Unlock()
	return client
}

// Deprecated: Use FetchV2
func (c *Client) Fetch(user *experiment.User) (map[string]experiment.Variant, error) {
	variants, err := c.FetchV2(user)
	if err != nil {
		return nil, err
	}
	results := filterDefaultVariants(variants)
	return results, nil
}

// FetchV2 fetches variants for a user from the remote evaluation service.
// Unlike Fetch, this method returns all variants, including default variants.
func (c *Client) FetchV2(user *experiment.User) (map[string]experiment.Variant, error) {
	ctx := context.Background()
	return c.FetchV2WithContext(user, ctx)
}

// FetchV2WithContext fetches variants for a user from the remote evaluation service with a context.
func (c *Client) FetchV2WithContext(user *experiment.User, ctx context.Context) (map[string]experiment.Variant, error) {
	return c.FetchV2WithContextAndOptions(user, ctx, nil)
}

// FetchV2WithOptions fetches variants for a user from the remote evaluation service with options.
func (c *Client) FetchV2WithOptions(user *experiment.User, fetchOptions *FetchOptions) (map[string]experiment.Variant, error) {
	ctx := context.Background()
	return c.FetchV2WithContextAndOptions(user, ctx, fetchOptions)
}

// FetchV2WithContextAndOptions fetches variants for a user from the remote evaluation service with a context and options.
func (c *Client) FetchV2WithContextAndOptions(user *experiment.User, ctx context.Context, fetchOptions *FetchOptions) (map[string]experiment.Variant, error) {
	variants, err := c.doFetch(ctx, user, c.config.FetchTimeout, fetchOptions)
	if err != nil {
		c.log.Error("fetch error: %v", err)
		if c.config.RetryBackoff.FetchRetries > 0 && shouldRetryFetch(err) {
			return c.retryFetch(ctx, user, fetchOptions)
		} else {
			return nil, err
		}
	}
	return variants, err
}

func (c *Client) doFetch(ctx context.Context, user *experiment.User, timeout time.Duration, fetchOptions *FetchOptions) (map[string]experiment.Variant, error) {
	addLibraryContext(user)
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/v2/vardata"
	if c.config.Debug {
		endpoint.RawQuery = fmt.Sprintf("d=%s", randStringRunes(5))
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	c.log.Debug("fetch variants for user %s", string(jsonBytes))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-User", base64.StdEncoding.EncodeToString(jsonBytes))
	if fetchOptions != nil {
		if fetchOptions.TracksAssignment {
			req.Header.Set("X-Amp-Exp-Track", "track")
		} else {
			req.Header.Set("X-Amp-Exp-Track", "no-track")
		}
		if fetchOptions.TracksExposure {
			req.Header.Set("X-Amp-Exp-Exposure-Track", "track")
		} else {
			req.Header.Set("X-Amp-Exp-Exposure-Track", "no-track")
		}
	}
	c.log.Debug("fetch request: %v", req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.log.Debug("fetch response: %v", *resp)
	if resp.StatusCode != http.StatusOK {
		return nil, &fetchError{StatusCode: resp.StatusCode, Message: resp.Status}
	}
	return c.parseResponse(resp)
}

func (c *Client) retryFetch(ctx context.Context, user *experiment.User, fetchOptions *FetchOptions) (map[string]experiment.Variant, error) {
	var err error
	var variants map[string]experiment.Variant
	var timer *time.Timer
	delay := c.config.RetryBackoff.FetchRetryBackoffMin
	for i := 0; i < c.config.RetryBackoff.FetchRetries; i++ {
		c.log.Debug("retry attempt %v", i)
		timer = time.NewTimer(delay)
		<-timer.C
		variants, err = c.doFetch(ctx, user, c.config.RetryBackoff.FetchRetryTimeout, fetchOptions)
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
	variants := make(map[string]experiment.Variant)
	err := json.NewDecoder(resp.Body).Decode(&variants)
	if err != nil {
		return nil, err
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
	if err, ok := err.(*fetchError); ok {
		return err.StatusCode < 400 || err.StatusCode >= 500 || err.StatusCode == 429
	}
	return true
}

func filterDefaultVariants(variants map[string]experiment.Variant) map[string]experiment.Variant {
	results := make(map[string]experiment.Variant)
	for key, variant := range variants {
		isDefault, ok := variant.Metadata["default"].(bool)
		if !ok {
			isDefault = false
		}
		isDeployed, ok := variant.Metadata["deployed"].(bool)
		if !ok {
			isDeployed = true
		}
		if !isDefault && isDeployed {
			results[key] = variant
		}
	}
	return results
}
