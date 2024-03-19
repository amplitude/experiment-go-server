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
	AssignmentConfig               *AssignmentConfig
	DeploymentKey                  string
}

type AssignmentConfig struct {
	amplitude.Config
	CacheCapacity int
}

var DefaultConfig = &Config{
	Debug:                          false,
	ServerUrl:                      "https://api.lab.amplitude.com/",
	FlagConfigPollerInterval:       30 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
}

var DefaultAssignmentConfig = &AssignmentConfig{
	CacheCapacity: 524288,
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
	if c.AssignmentConfig != nil && c.AssignmentConfig.CacheCapacity == 0 {
		c.AssignmentConfig.CacheCapacity = DefaultAssignmentConfig.CacheCapacity
	}
	return c
}
