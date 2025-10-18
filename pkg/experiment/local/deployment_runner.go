package local

import (
	"sync"
	"time"
)

type deploymentRunner struct {
	config            *Config
	flagConfigStorage flagConfigStorage
	flagConfigUpdater flagConfigUpdater
	cohortLoader      *cohortLoader
	poller            *poller
	lock              sync.Mutex
}

const streamUpdaterRetryDelay = 15 * time.Second
const updaterRetryMaxJitter = 1 * time.Second

func newDeploymentRunner(
	config *Config,
	flagConfigApi flagConfigApi,
	flagConfigStreamApi *flagConfigStreamApiV2,
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
) *deploymentRunner {
	flagConfigUpdater := newflagConfigFallbackRetryWrapper(newFlagConfigPoller(flagConfigApi, config, flagConfigStorage, cohortStorage, cohortLoader), nil, config.FlagConfigPollerInterval, updaterRetryMaxJitter, 0, 0, config.LogLevel, config.LoggerProvider)
	if flagConfigStreamApi != nil {
		flagConfigUpdater = newflagConfigFallbackRetryWrapper(newFlagConfigStreamer(flagConfigStreamApi, config, flagConfigStorage, cohortStorage, cohortLoader), flagConfigUpdater, streamUpdaterRetryDelay, updaterRetryMaxJitter, config.FlagConfigPollerInterval, 0, config.LogLevel, config.LoggerProvider)
	}
	dr := &deploymentRunner{
		config:            config,
		flagConfigStorage: flagConfigStorage,
		cohortLoader:      cohortLoader,
		flagConfigUpdater: flagConfigUpdater,
		poller:            newPoller(),
	}
	return dr
}

func (dr *deploymentRunner) start() error {
	dr.lock.Lock()
	defer dr.lock.Unlock()
	err := dr.flagConfigUpdater.Start(nil)
	if err != nil {
		return err
	}

	if dr.config.CohortSyncConfig != nil {
		dr.poller.Poll(dr.config.CohortSyncConfig.CohortPollingInterval, func() {
			cohortIDs := getAllCohortIDsFromFlags(dr.flagConfigStorage.getFlagConfigsArray())
			dr.cohortLoader.downloadCohorts(cohortIDs)
		})
	}
	return nil
}
