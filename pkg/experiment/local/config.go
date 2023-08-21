package local

import (
	"github.com/amplitude/analytics-go/amplitude"
	"time"
)

type Config struct {
	Debug                          bool
	ServerUrl                      string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
	AssignmentConfig               AssignmentConfig
}

type AssignmentConfig struct {
	ApiKey         string
	FilterCapacity int
	AmpConfig      amplitude.Config
}

var DefaultConfig = &Config{
	Debug:                          false,
	ServerUrl:                      "https://api.lab.amplitude.com/",
	FlagConfigPollerInterval:       30 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
}

var DefaultAssignmentConfig = &AssignmentConfig{
	FilterCapacity: 65536,
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
	if c.AssignmentConfig.FilterCapacity == 0 {
		c.AssignmentConfig.FilterCapacity = DefaultAssignmentConfig.FilterCapacity
	}
	return c
}
