package local

import "time"

type Config struct {
	Debug                          bool
	ServerUrl                      string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
}

var DefaultConfig = &Config{
	Debug:                          false,
	ServerUrl:                      "https://api.lab.amplitude.com/",
	FlagConfigPollerInterval:       30 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
}

func fillConfigDefaults(c *Config) *Config {
	if c == nil {
		return DefaultConfig
	}
	if c.ServerUrl == "" {
		c.ServerUrl = DefaultConfig.ServerUrl
	}
	if c.FlagConfigPollerInterval == 0 {
		c.FlagConfigPollerInterval = DefaultConfig.FlagConfigPollerInterval
	}
	if c.FlagConfigPollerRequestTimeout == 0 {
		c.FlagConfigPollerRequestTimeout = DefaultConfig.FlagConfigPollerRequestTimeout
	}
	return c
}
