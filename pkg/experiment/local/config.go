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
	CohortSyncConfig               *CohortSyncConfig
}

type AssignmentConfig struct {
	amplitude.Config
	CacheCapacity int
}

// CohortSyncConfig holds configuration for cohort synchronization.
type CohortSyncConfig struct {
	ApiKey                   string
	SecretKey                string
	MaxCohortSize            int
	CohortRequestDelayMillis int
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

var DefaultCohortSyncConfig = &CohortSyncConfig{
	MaxCohortSize:            15000,
	CohortRequestDelayMillis: 5000,
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

	if c.CohortSyncConfig != nil && c.CohortSyncConfig.MaxCohortSize == 0 {
		c.CohortSyncConfig.MaxCohortSize = DefaultCohortSyncConfig.MaxCohortSize
	}

	if c.CohortSyncConfig != nil && c.CohortSyncConfig.CohortRequestDelayMillis == 0 {
		c.CohortSyncConfig.CohortRequestDelayMillis = DefaultCohortSyncConfig.CohortRequestDelayMillis
	}

	return c
}
