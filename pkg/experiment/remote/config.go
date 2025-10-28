package remote

import (
	"time"

	"github.com/amplitude/experiment-go-server/pkg/logger"
)

type Config struct {
	Debug        		bool
	LogLevel				logger.LogLevel
	LoggerProvider	logger.LoggerProvider
	ServerUrl    		string
	FetchTimeout 		time.Duration
	RetryBackoff 		*RetryBackoff
}

var DefaultConfig = &Config{
	Debug:        	false,
	LogLevel:				logger.Error,
	LoggerProvider:	logger.NewDefault(),
	ServerUrl:    	"https://api.lab.amplitude.com/",
	FetchTimeout: 	500 * time.Millisecond,
	RetryBackoff: 	DefaultRetryBackoff,
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
	if c.LogLevel == logger.Unknown {
		if c.Debug {
			c.LogLevel = logger.Debug
		} else {
			c.LogLevel = logger.Error
		}
	}
	return c
}
