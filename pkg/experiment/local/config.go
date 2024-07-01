package local

import (
	"github.com/amplitude/analytics-go/amplitude"
	"math"
	"strings"
	"time"
)

const EUFlagServerUrl = "https://flag.lab.eu.amplitude.com"
const EUCohortSyncUrl = "https://cohort-v2.lab.eu.amplitude.com"

type Config struct {
	Debug                          bool
	ServerUrl                      string
	ServerZone                     string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
	AssignmentConfig               *AssignmentConfig
	CohortSyncConfig               *CohortSyncConfig
}

type AssignmentConfig struct {
	amplitude.Config
	CacheCapacity int
}

type CohortSyncConfig struct {
	ApiKey                   string
	SecretKey                string
	MaxCohortSize            int
	CohortRequestDelayMillis int
	CohortServerUrl          string
}

var DefaultConfig = &Config{
	Debug:                          false,
	ServerUrl:                      "https://api.lab.amplitude.com/",
	ServerZone:                     "us",
	FlagConfigPollerInterval:       30 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
}

var DefaultAssignmentConfig = &AssignmentConfig{
	CacheCapacity: 524288,
}

var DefaultCohortSyncConfig = &CohortSyncConfig{
	MaxCohortSize:            math.MaxInt32,
	CohortRequestDelayMillis: 5000,
	CohortServerUrl:          "https://cohort-v2.lab.amplitude.com",
}

func fillConfigDefaults(c *Config) *Config {
	if c == nil {
		return DefaultConfig
	}
	if c.ServerZone == "" {
		c.ServerZone = DefaultConfig.ServerZone
	}
	if c.ServerUrl == "" {
		if strings.EqualFold(strings.ToLower(c.ServerZone), strings.ToLower(DefaultConfig.ServerZone)) {
			c.ServerUrl = DefaultConfig.ServerUrl
		} else if strings.EqualFold(strings.ToLower(c.ServerZone), "eu") {
			c.ServerUrl = EUFlagServerUrl
		}
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

	if c.CohortSyncConfig != nil && c.CohortSyncConfig.CohortServerUrl == "" {
		if strings.EqualFold(strings.ToLower(c.ServerZone), strings.ToLower(DefaultConfig.ServerZone)) {
			c.CohortSyncConfig.CohortServerUrl = DefaultCohortSyncConfig.CohortServerUrl
		} else if strings.EqualFold(strings.ToLower(c.ServerZone), "eu") {
			c.CohortSyncConfig.CohortServerUrl = EUCohortSyncUrl
		}
	}

	return c
}
