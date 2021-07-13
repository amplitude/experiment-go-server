package experiment

import "time"

type Config struct {
	Debug        bool
	ServerUrl    string
	FetchTimeout time.Duration
	RetryBackoff *RetryBackoff
}

type RetryBackoff struct {
	FetchRetries            int
	FetchRetryBackoffMin    time.Duration
	FetchRetryBackoffMax    time.Duration
	FetchRetryBackoffScalar float64
	FetchRetryTimeout       time.Duration
}

var DefaultRetryBackoff = &RetryBackoff{
	FetchRetries:            8,
	FetchRetryBackoffMin:    500 * time.Millisecond,
	FetchRetryBackoffMax:    10_000 * time.Millisecond,
	FetchRetryBackoffScalar: 1.5,
	FetchRetryTimeout:       10_000 * time.Millisecond,
}

var DefaultConfig = &Config{
	Debug:        false,
	ServerUrl:    "https://api.lab.amplitude.com/",
	FetchTimeout: 10_000 * time.Millisecond,
	RetryBackoff: DefaultRetryBackoff,
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
