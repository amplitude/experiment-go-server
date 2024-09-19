package local

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/internal/logger"
)

type flagConfigUpdater interface {
	Start(func (error)) error
	Stop()
}

type flagConfigUpdaterBase struct {
	flagConfigStorage flagConfigStorage
	cohortStorage cohortStorage
	cohortLoader *cohortLoader
	log               *logger.Log
}

func newFlagConfigUpdaterBase(
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
	config *Config,
) flagConfigUpdaterBase {
	return flagConfigUpdaterBase{
		flagConfigStorage: flagConfigStorage,
		cohortStorage: cohortStorage,
		cohortLoader: cohortLoader,
		log: logger.New(config.Debug),
	}
}

func (dr *flagConfigUpdaterBase) update(flagConfigs map[string]*evaluation.Flag) error {
	
	flagKeys := make(map[string]struct{})
	for _, flag := range flagConfigs {
		flagKeys[flag.Key] = struct{}{}
	}

	dr.flagConfigStorage.removeIf(func(f *evaluation.Flag) bool {
		_, exists := flagKeys[f.Key]
		return !exists
	})

	if dr.cohortLoader == nil {
		for _, flagConfig := range flagConfigs {
			dr.log.Debug("Putting non-cohort flag %s", flagConfig.Key)
			dr.flagConfigStorage.putFlagConfig(flagConfig)
		}
		return nil
	}

	newCohortIDs := make(map[string]struct{})
	for _, flagConfig := range flagConfigs {
		for cohortID := range getAllCohortIDsFromFlag(flagConfig) {
			newCohortIDs[cohortID] = struct{}{}
		}
	}

	existingCohortIDs := dr.cohortStorage.getCohortIds()
	cohortIDsToDownload := difference(newCohortIDs, existingCohortIDs)

	// Download all new cohorts
	dr.cohortLoader.downloadCohorts(cohortIDsToDownload)

	// Get updated set of cohort ids
	updatedCohortIDs := dr.cohortStorage.getCohortIds()
	// Iterate through new flag configs and check if their required cohorts exist
	for _, flagConfig := range flagConfigs {
		cohortIDs := getAllCohortIDsFromFlag(flagConfig)
		missingCohorts := difference(cohortIDs, updatedCohortIDs)

		dr.flagConfigStorage.putFlagConfig(flagConfig)
		dr.log.Debug("Putting flag %s", flagConfig.Key)
		if len(missingCohorts) != 0 {
			dr.log.Error("Flag %s - failed to load cohorts: %v", flagConfig.Key, missingCohorts)
		}
	}

	// Delete unused cohorts
	dr.deleteUnusedCohorts()
	dr.log.Debug("Refreshed %d flag configs.", len(flagConfigs))

	return nil
}

func (dr *flagConfigUpdaterBase) deleteUnusedCohorts() {
	flagCohortIDs := make(map[string]struct{})
	for _, flag := range dr.flagConfigStorage.getFlagConfigs() {
		for cohortID := range getAllCohortIDsFromFlag(flag) {
			flagCohortIDs[cohortID] = struct{}{}
		}
	}

	storageCohorts := dr.cohortStorage.getCohorts()
	for cohortID := range storageCohorts {
		if _, exists := flagCohortIDs[cohortID]; !exists {
			cohort := storageCohorts[cohortID]
			if cohort != nil {
				dr.cohortStorage.deleteCohort(cohort.GroupType, cohortID)
			}
		}
	}
}

type flagConfigStreamer struct {
	flagConfigUpdaterBase
    flagConfigStreamApi *flagConfigStreamApiV2
	lock sync.Mutex
}

func NewFlagConfigStreamer(
    flagConfigStreamApi *flagConfigStreamApiV2,
    config *Config,
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
) flagConfigUpdater {
	return &flagConfigStreamer{
		flagConfigStreamApi: flagConfigStreamApi,
		flagConfigUpdaterBase: newFlagConfigUpdaterBase(flagConfigStorage, cohortStorage, cohortLoader, config),
	}
}

func (a *flagConfigStreamer) Start(onError func (error)) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.stopInternal()
	a.flagConfigStreamApi.OnUpdate = func (flags map[string]*evaluation.Flag) error {
		fmt.Println("stream got data")
		return a.update(flags)
	}
	a.flagConfigStreamApi.OnError = func (err error) {
		fmt.Println("stream got err", err)
		if (onError != nil) {
			onError(err)
		}
	}
	return a.flagConfigStreamApi.Connect()
}

func (a *flagConfigStreamer) stopInternal() {
	a.flagConfigStreamApi.Close()
}

func (a *flagConfigStreamer) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.stopInternal()
}

type flagConfigPoller struct {
	flagConfigUpdaterBase
    flagConfigApi flagConfigApi
    config *Config
	poller            *poller
	lock sync.Mutex
}

func NewFlagConfigPoller(
    flagConfigApi flagConfigApi,
    config *Config,
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
) flagConfigUpdater {
	return &flagConfigPoller{
		flagConfigApi: flagConfigApi,
		config: config,
		flagConfigUpdaterBase: newFlagConfigUpdaterBase(flagConfigStorage, cohortStorage, cohortLoader, config),
	}
}

func (a *flagConfigPoller) Start(onError func (error)) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if err := a.stopInternal(); err != nil {
		return err
	}

	if err := a.updateFlagConfigs(); err != nil {
		a.log.Error("Initial updateFlagConfigs failed: %v", err)
		return err
	}

	a.poller = newPoller()
	a.poller.Poll(a.config.FlagConfigPollerInterval, func() {
		if err := a.periodicRefresh(); err != nil {
			a.log.Error("Periodic updateFlagConfigs failed: %v", err)
			a.Stop()
			onError(err)
		}
	})
	return nil
}

func (dr *flagConfigPoller) periodicRefresh() error {
	defer func() {
		if r := recover(); r != nil {
			dr.log.Error("Recovered in periodicRefresh: %v", r)
		}
	}()
	return dr.updateFlagConfigs()
}

func (a *flagConfigPoller) updateFlagConfigs() error {
	a.log.Debug("Refreshing flag configs.")
	flagConfigs, err := a.flagConfigApi.getFlagConfigs()
	if err != nil {
		a.log.Error("Failed to fetch flag configs: %v", err)
		return err
	}

	return a.update(flagConfigs)
}

func (a *flagConfigPoller) stopInternal() error {
	if (a.poller != nil) {
		close(a.poller.shutdown)
		a.poller = nil
	}
	return nil
}

func (a *flagConfigPoller) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.stopInternal()
}

type FlagConfigFallbackRetryWrapper struct {
    mainUpdater flagConfigUpdater
    fallbackUpdater flagConfigUpdater
    retryDelay time.Duration
    maxJitter time.Duration
	retryCancelCh chan bool
	lock sync.Mutex
}
func NewFlagConfigFallbackRetryWrapper(
    mainUpdater flagConfigUpdater,
    fallbackUpdater flagConfigUpdater,
    retryDelay time.Duration,
    maxJitter time.Duration,
) flagConfigUpdater {
	return &FlagConfigFallbackRetryWrapper{
		mainUpdater: mainUpdater,
		fallbackUpdater: fallbackUpdater,
		retryDelay: retryDelay,
		maxJitter: maxJitter,
	}
}

    /**
     * Since the wrapper retries, so there will never be error case. Thus, onError will never be called.
     */
func (a *FlagConfigFallbackRetryWrapper) Start(onError func (error)) error {
	// if (mainUpdater is FlagConfigFallbackRetryWrapper) {
	//     throw Error("Do not use FlagConfigFallbackRetryWrapper as main updater. Fallback updater will never be used. Rewrite retry and fallback logic.")
	// }

	a.lock.Lock()
	defer a.lock.Unlock()

	if (a.retryCancelCh != nil) {
		close(a.retryCancelCh)
		a.retryCancelCh = nil
	}

	err := a.mainUpdater.Start(func (error) {
		a.scheduleRetry() // Don't care if poller start error or not, always retry.
		if (a.fallbackUpdater != nil) {
			a.fallbackUpdater.Start(nil)
		}
	})
	if (err == nil) {
		fmt.Println("main start ok")
		// Main start success, stop fallback.
		if (a.fallbackUpdater != nil) {
			a.fallbackUpdater.Stop()
		}
		return nil
	}
	fmt.Println("main start err", err)
	// Logger.e("Primary flag configs start failed, start fallback. Error: ", t)
	if (a.fallbackUpdater == nil) {
		// No fallback, main start failed is wrapper start fail
		return err
	}
	err = a.fallbackUpdater.Start(nil)
	if (err != nil) {
		return err
	}

	go func() {a.scheduleRetry()}()
	return nil
}

func (a *FlagConfigFallbackRetryWrapper) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if (a.retryCancelCh != nil) {
		close(a.retryCancelCh)
		a.retryCancelCh = nil
	}
	a.mainUpdater.Stop()
	if (a.fallbackUpdater != nil) {
		a.fallbackUpdater.Stop()
	}
}

func (a *FlagConfigFallbackRetryWrapper) scheduleRetry() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if (a.retryCancelCh != nil) {
		close(a.retryCancelCh)
	}
	a.retryCancelCh = make(chan bool)
	go func() {
		select {
		case <-a.retryCancelCh: return
		case <-time.After(a.retryDelay - a.maxJitter + time.Duration(rand.Int64N(a.maxJitter.Nanoseconds() * 2))):
			fmt.Println("retrying")
			a.lock.Lock()
			defer a.lock.Unlock()

			close(a.retryCancelCh)
			a.retryCancelCh = nil

			err := a.mainUpdater.Start(func (error) {
				a.scheduleRetry() // Don't care if poller start error or not, always retry.
				if (a.fallbackUpdater != nil) {
					a.fallbackUpdater.Start(nil)
				}
			})
			if (err == nil) {
				// Main start success, stop fallback.
				if (a.fallbackUpdater != nil) {
					a.fallbackUpdater.Stop()
				}
				return
			}
			
			go func() {a.scheduleRetry()}()
		}
	}()
}