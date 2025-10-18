package local

import (
	"math"
	"time"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/amplitude/experiment-go-server/logger"
)

const EUFlagServerUrl = "https://flag.lab.eu.amplitude.com"
const EUFlagStreamServerUrl = "https://stream.lab.eu.amplitude.com"
const EUCohortSyncUrl = "https://cohort-v2.lab.eu.amplitude.com"

type ServerZone int

const (
	USServerZone ServerZone = iota
	EUServerZone
)

type Config struct {
	Debug                          bool
	LogLevel 										   logger.LogLevel											 
	LoggerProvider								 logger.LoggerProvider
	ServerUrl                      string
	ServerZone                     ServerZone
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
	StreamUpdates                  bool
	StreamServerUrl                string
	StreamFlagConnTimeout          time.Duration
	AssignmentConfig               *AssignmentConfig
	CohortSyncConfig               *CohortSyncConfig
}

type AssignmentConfig struct {
	amplitude.Config
	CacheCapacity int
}

type CohortSyncConfig struct {
	ApiKey                string
	SecretKey             string
	MaxCohortSize         int
	CohortPollingInterval time.Duration
	CohortServerUrl       string
}

var DefaultConfig = &Config{
	Debug:                          false,
	LogLevel:												logger.Error,
	LoggerProvider:									logger.NewDefault(),
	ServerUrl:                      "https://api.lab.amplitude.com/",
	ServerZone:                     USServerZone,
	FlagConfigPollerInterval:       30 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
	StreamUpdates:                  false,
	StreamServerUrl:                "https://stream.lab.amplitude.com",
	StreamFlagConnTimeout:          1500 * time.Millisecond,
}

var DefaultAssignmentConfig = &AssignmentConfig{
	CacheCapacity: 524288,
}

var DefaultCohortSyncConfig = &CohortSyncConfig{
	MaxCohortSize:         math.MaxInt32,
	CohortPollingInterval: 60 * time.Second,
	CohortServerUrl:       "https://cohort-v2.lab.amplitude.com",
}

func fillConfigDefaults(c *Config) *Config {
	if c == nil {
		return DefaultConfig
	}
	if c.ServerZone == 0 {
		c.ServerZone = DefaultConfig.ServerZone
	}
	if c.ServerUrl == "" {
		switch c.ServerZone {
		case USServerZone:
			c.ServerUrl = DefaultConfig.ServerUrl
			c.StreamServerUrl = DefaultConfig.StreamServerUrl
		case EUServerZone:
			c.ServerUrl = EUFlagServerUrl
			c.StreamServerUrl = EUFlagStreamServerUrl
		}
	}

	if c.FlagConfigPollerInterval == 0 {
		c.FlagConfigPollerInterval = DefaultConfig.FlagConfigPollerInterval
	}
	if c.FlagConfigPollerRequestTimeout == 0 {
		c.FlagConfigPollerRequestTimeout = DefaultConfig.FlagConfigPollerRequestTimeout
	}
	if c.StreamFlagConnTimeout == 0 {
		c.StreamFlagConnTimeout = DefaultConfig.StreamFlagConnTimeout
	}
	if c.AssignmentConfig != nil && c.AssignmentConfig.CacheCapacity == 0 {
		c.AssignmentConfig.CacheCapacity = DefaultAssignmentConfig.CacheCapacity
	}

	if c.CohortSyncConfig != nil && c.CohortSyncConfig.MaxCohortSize == 0 {
		c.CohortSyncConfig.MaxCohortSize = DefaultCohortSyncConfig.MaxCohortSize
	}

	if c.CohortSyncConfig != nil && (c.CohortSyncConfig.CohortPollingInterval < DefaultCohortSyncConfig.CohortPollingInterval) {
		c.CohortSyncConfig.CohortPollingInterval = DefaultCohortSyncConfig.CohortPollingInterval
	}

	if c.CohortSyncConfig != nil && c.CohortSyncConfig.CohortServerUrl == "" {
		switch c.ServerZone {
		case USServerZone:
			c.CohortSyncConfig.CohortServerUrl = DefaultCohortSyncConfig.CohortServerUrl
		case EUServerZone:
			c.CohortSyncConfig.CohortServerUrl = EUCohortSyncUrl
		}
	}

	if c.LogLevel == logger.Unknown {
		c.LogLevel = logger.Error
	}

	if c.Debug {
		c.LogLevel = logger.Debug
	} 

	if c.LoggerProvider == nil {
		c.LoggerProvider = logger.NewDefault()
	}

	return c
}
