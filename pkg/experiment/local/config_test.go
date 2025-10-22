package local

import (
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/logger"
)

func TestFillConfigDefaults_ServerZoneAndServerUrl(t *testing.T) {
	tests := []struct {
		name         string
		input        *Config
		expectedZone ServerZone
		expectedUrl  string
		expectedStreamUrl  string
	}{
		{
			name:         "Nil config",
			input:        nil,
			expectedZone: DefaultConfig.ServerZone,
			expectedUrl:  DefaultConfig.ServerUrl,
			expectedStreamUrl:  DefaultConfig.StreamServerUrl,
		},
		{
			name:         "Empty ServerZone",
			input:        &Config{},
			expectedZone: DefaultConfig.ServerZone,
			expectedUrl:  DefaultConfig.ServerUrl,
			expectedStreamUrl:  DefaultConfig.StreamServerUrl,
		},
		{
			name:         "ServerZone US",
			input:        &Config{ServerZone: USServerZone},
			expectedZone: USServerZone,
			expectedUrl:  DefaultConfig.ServerUrl,
			expectedStreamUrl:  DefaultConfig.StreamServerUrl,
		},
		{
			name:         "ServerZone EU",
			input:        &Config{ServerZone: EUServerZone},
			expectedZone: EUServerZone,
			expectedUrl:  EUFlagServerUrl,
			expectedStreamUrl:  EUFlagStreamServerUrl,
		},
		{
			name:         "Uppercase ServerZone EU",
			input:        &Config{ServerZone: EUServerZone},
			expectedZone: EUServerZone,
			expectedUrl:  EUFlagServerUrl,
			expectedStreamUrl:  EUFlagStreamServerUrl,
		},
		{
			name:         "Custom ServerUrl",
			input:        &Config{ServerZone: USServerZone, ServerUrl: "https://custom.url/", StreamServerUrl: "https://stream.custom.url"},
			expectedZone: USServerZone,
			expectedUrl:  "https://custom.url/",
			expectedStreamUrl:  "https://stream.custom.url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillConfigDefaults(tt.input)
			if result.ServerZone != tt.expectedZone {
				t.Errorf("expected ServerZone %d, got %d", tt.expectedZone, result.ServerZone)
			}
			if result.ServerUrl != tt.expectedUrl {
				t.Errorf("expected ServerUrl %s, got %s", tt.expectedUrl, result.ServerUrl)
			}
			if result.StreamServerUrl != tt.expectedStreamUrl {
				t.Errorf("expected StreamServerUrl %s, got %s", tt.expectedStreamUrl, result.StreamServerUrl)
			}
		})
	}
}

func TestFillConfigDefaults_CohortSyncConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       *Config
		expectedUrl string
	}{
		{
			name: "Nil CohortSyncConfig",
			input: &Config{
				ServerZone:       EUServerZone,
				CohortSyncConfig: nil,
			},
			expectedUrl: "",
		},
		{
			name: "CohortSyncConfig with empty CohortServerUrl",
			input: &Config{
				ServerZone:       EUServerZone,
				CohortSyncConfig: &CohortSyncConfig{},
			},
			expectedUrl: EUCohortSyncUrl,
		},
		{
			name: "CohortSyncConfig with custom CohortServerUrl",
			input: &Config{
				ServerZone: USServerZone,
				CohortSyncConfig: &CohortSyncConfig{
					CohortServerUrl: "https://custom-cohort.url/",
				},
			},
			expectedUrl: "https://custom-cohort.url/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillConfigDefaults(tt.input)
			if tt.input.CohortSyncConfig == nil {
				if result.CohortSyncConfig == nil {
					return
				}
				if result.CohortSyncConfig.CohortServerUrl != tt.expectedUrl {
					t.Errorf("expected CohortServerUrl %s, got %s", tt.expectedUrl, result.CohortSyncConfig.CohortServerUrl)
				}
			} else {
				if result.CohortSyncConfig.CohortServerUrl != tt.expectedUrl {
					t.Errorf("expected CohortServerUrl %s, got %s", tt.expectedUrl, result.CohortSyncConfig.CohortServerUrl)
				}
			}
		})
	}
}

func TestFillConfigDefaults_DefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name:     "Nil config",
			input:    nil,
			expected: DefaultConfig,
		},
		{
			name:  "Empty config",
			input: &Config{},
			expected: &Config{
				ServerZone:                     DefaultConfig.ServerZone,
				ServerUrl:                      DefaultConfig.ServerUrl,
				StreamServerUrl:                DefaultConfig.StreamServerUrl,
				FlagConfigPollerInterval:       DefaultConfig.FlagConfigPollerInterval,
				FlagConfigPollerRequestTimeout: DefaultConfig.FlagConfigPollerRequestTimeout,
				LogLevel: 											DefaultConfig.LogLevel,
				LoggerProvider: 								DefaultConfig.LoggerProvider,
			},
		},
		{
			name: "Custom values",
			input: &Config{
				ServerZone:                     EUServerZone,
				ServerUrl:                      "https://custom.url/",
				StreamServerUrl:                "https://stream.custom.url",
				FlagConfigPollerInterval:       60 * time.Second,
				FlagConfigPollerRequestTimeout: 20 * time.Second,
				LogLevel: 											logger.Info,
			},
			expected: &Config{
				ServerZone:                     EUServerZone,
				ServerUrl:                      "https://custom.url/",
				StreamServerUrl:                "https://stream.custom.url",
				FlagConfigPollerInterval:       60 * time.Second,
				FlagConfigPollerRequestTimeout: 20 * time.Second,
				LogLevel:												logger.Info,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillConfigDefaults(tt.input)
			if result.ServerZone != tt.expected.ServerZone {
				t.Errorf("expected ServerZone %d, got %d", tt.expected.ServerZone, result.ServerZone)
			}
			if result.ServerUrl != tt.expected.ServerUrl {
				t.Errorf("expected ServerUrl %s, got %s", tt.expected.ServerUrl, result.ServerUrl)
			}
			if result.StreamServerUrl != tt.expected.StreamServerUrl {
				t.Errorf("expected StreamServerUrl %s, got %s", tt.expected.StreamServerUrl, result.StreamServerUrl)
			}
			if result.FlagConfigPollerInterval != tt.expected.FlagConfigPollerInterval {
				t.Errorf("expected FlagConfigPollerInterval %v, got %v", tt.expected.FlagConfigPollerInterval, result.FlagConfigPollerInterval)
			}
			if result.FlagConfigPollerRequestTimeout != tt.expected.FlagConfigPollerRequestTimeout {
				t.Errorf("expected FlagConfigPollerRequestTimeout %v, got %v", tt.expected.FlagConfigPollerRequestTimeout, result.FlagConfigPollerRequestTimeout)
			}
			if result.LogLevel != tt.expected.LogLevel {
				t.Errorf("expected LogLevel %d, got %d", tt.expected.LogLevel, result.LogLevel)
			}
		})
	}
}

func TestFillLogDefaults_LoggingValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name:     "Nil config",
			input:    nil,
			expected: DefaultConfig,
		},
		{
			name:  "Empty config with Debug true",
			input: &Config{
				Debug:													true,
			},
			expected: &Config{
				LogLevel: 											logger.Debug,
			},
		},
		{
			name: "Custom level with Debug true",
			input: &Config{
				Debug:													true,
				LogLevel: 											logger.Info,
			},
			expected: &Config{
				LogLevel: 											logger.Debug,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillConfigDefaults(tt.input)
			if result.LogLevel != tt.expected.LogLevel {
				t.Errorf("expected LogLevel %d, got %d", tt.expected.LogLevel, result.LogLevel)
			}
		})
	}
}
