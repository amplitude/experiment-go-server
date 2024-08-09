package remote

import (
	"net/http"
	"time"
)

type Config struct {
	Debug        bool
	ServerUrl    string
	FetchTimeout time.Duration
	RetryBackoff *RetryBackoff
	HttpClient   *http.Client
}

var DefaultConfig = &Config{
	Debug:        false,
	ServerUrl:    "https://api.lab.amplitude.com/",
	FetchTimeout: 500 * time.Millisecond,
	RetryBackoff: DefaultRetryBackoff,
	HttpClient:   nil,
}

type RetryBackoff struct {
	FetchRetries            int
	FetchRetryBackoffMin    time.Duration
	FetchRetryBackoffMax    time.Duration
	FetchRetryBackoffScalar float64
	FetchRetryTimeout       time.Duration
}

var DefaultRetryBackoff = &RetryBackoff{
	FetchRetries:            1,
	FetchRetryBackoffMin:    0 * time.Millisecond,
	FetchRetryBackoffMax:    10000 * time.Millisecond,
	FetchRetryBackoffScalar: 1,
	FetchRetryTimeout:       500 * time.Millisecond,
}

func fillConfigDefaults(c *Config) *Config {
	if c == nil {
		return DefaultConfig
	}
	if c.ServerUrl == "" {
		c.ServerUrl = DefaultConfig.ServerUrl
	}
	if c.FetchTimeout == 0 {
		c.FetchTimeout = DefaultConfig.FetchTimeout
	}
	if c.RetryBackoff == nil {
		c.RetryBackoff = DefaultConfig.RetryBackoff
	}
	return c
}
