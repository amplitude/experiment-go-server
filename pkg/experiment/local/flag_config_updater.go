package local

import (
	"fmt"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/internal/logger"
)

type flagConfigUpdater interface {
    // Start the updater. There can be multiple calls.
    // If start fails, it should return err. The caller should handle error.
    // If some other async error happened while updating (after already started successfully),
	//   it should call the `func (error)` callback function.
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

func (u *flagConfigUpdaterBase) update(flagConfigs map[string]*evaluation.Flag) error {
	
	flagKeys := make(map[string]struct{})
	for _, flag := range flagConfigs {
		flagKeys[flag.Key] = struct{}{}
	}

	u.flagConfigStorage.removeIf(func(f *evaluation.Flag) bool {
		_, exists := flagKeys[f.Key]
		return !exists
	})

	if u.cohortLoader == nil {
		for _, flagConfig := range flagConfigs {
			u.log.Debug("Putting non-cohort flag %s", flagConfig.Key)
			u.flagConfigStorage.putFlagConfig(flagConfig)
		}
		return nil
	}

	newCohortIDs := make(map[string]struct{})
	for _, flagConfig := range flagConfigs {
		for cohortID := range getAllCohortIDsFromFlag(flagConfig) {
			newCohortIDs[cohortID] = struct{}{}
		}
	}

	existingCohortIDs := u.cohortStorage.getCohortIds()
	cohortIDsToDownload := difference(newCohortIDs, existingCohortIDs)

	// Download all new cohorts
	u.cohortLoader.downloadCohorts(cohortIDsToDownload)

	// Get updated set of cohort ids
	updatedCohortIDs := u.cohortStorage.getCohortIds()
	// Iterate through new flag configs and check if their required cohorts exist
	for _, flagConfig := range flagConfigs {
		cohortIDs := getAllCohortIDsFromFlag(flagConfig)
		missingCohorts := difference(cohortIDs, updatedCohortIDs)

		u.flagConfigStorage.putFlagConfig(flagConfig)
		u.log.Debug("Putting flag %s", flagConfig.Key)
		if len(missingCohorts) != 0 {
			u.log.Error("Flag %s - failed to load cohorts: %v", flagConfig.Key, missingCohorts)
		}
	}

	// Delete unused cohorts
	u.deleteUnusedCohorts()
	u.log.Debug("Refreshed %d flag configs.", len(flagConfigs))

	return nil
}

func (u *flagConfigUpdaterBase) deleteUnusedCohorts() {
	flagCohortIDs := make(map[string]struct{})
	for _, flag := range u.flagConfigStorage.getFlagConfigs() {
		for cohortID := range getAllCohortIDsFromFlag(flag) {
			flagCohortIDs[cohortID] = struct{}{}
		}
	}

	storageCohorts := u.cohortStorage.getCohorts()
	for cohortID := range storageCohorts {
		if _, exists := flagCohortIDs[cohortID]; !exists {
			cohort := storageCohorts[cohortID]
			if cohort != nil {
				u.cohortStorage.deleteCohort(cohort.GroupType, cohortID)
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

func (s *flagConfigStreamer) Start(onError func (error)) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stopInternal()
	return s.flagConfigStreamApi.Connect(
		nil, 
		func (flags map[string]*evaluation.Flag) error {
			return s.update(flags)
		},
		func (err error) {
			if (onError != nil) {
				onError(err)
			}
		},
	)
}

func (s *flagConfigStreamer) stopInternal() {
	s.flagConfigStreamApi.Close()
}

func (s *flagConfigStreamer) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stopInternal()
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

func (p *flagConfigPoller) Start(onError func (error)) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err := p.stopInternal(); err != nil {
		return err
	}

	if err := p.updateFlagConfigs(); err != nil {
		p.log.Error("Initial updateFlagConfigs failed: %v", err)
		return err
	}

	p.poller = newPoller()
	p.poller.Poll(p.config.FlagConfigPollerInterval, func() {
		if err := p.periodicRefresh(); err != nil {
			p.log.Error("Periodic updateFlagConfigs failed: %v", err)
			p.Stop()
			onError(err)
		}
	})
	return nil
}

func (p *flagConfigPoller) periodicRefresh() error {
	defer func() {
		if r := recover(); r != nil {
			p.log.Error("Recovered in periodicRefresh: %v", r)
		}
	}()
	return p.updateFlagConfigs()
}

func (p *flagConfigPoller) updateFlagConfigs() error {
	p.log.Debug("Refreshing flag configs.")
	flagConfigs, err := p.flagConfigApi.getFlagConfigs()
	if err != nil {
		p.log.Error("Failed to fetch flag configs: %v", err)
		return err
	}

	return p.update(flagConfigs)
}

func (p *flagConfigPoller) stopInternal() error {
	if (p.poller != nil) {
		close(p.poller.shutdown)
		p.poller = nil
	}
	return nil
}

func (p *flagConfigPoller) Stop() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.stopInternal()
}

type FlagConfigFallbackRetryWrapper struct {
    mainUpdater flagConfigUpdater
    fallbackUpdater flagConfigUpdater
    retryDelay time.Duration
    maxJitter time.Duration
	retryTimer *time.Timer
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
func (w *FlagConfigFallbackRetryWrapper) Start(onError func (error)) error {
	// if (mainUpdater is FlagConfigFallbackRetryWrapper) {
	//     throw Error("Do not use FlagConfigFallbackRetryWrapper as main updater. Fallback updater will never be used. Rewrite retry and fallback logic.")
	// }

	w.lock.Lock()
	defer w.lock.Unlock()

	if (w.retryTimer != nil) {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}

	err := w.mainUpdater.Start(func (error) {
		go func() {w.scheduleRetry()}() // Don't care if poller start error or not, always retry.
		if (w.fallbackUpdater != nil) {
			w.fallbackUpdater.Start(nil)
		}
	})
	if (err == nil) {
		fmt.Println("main start ok")
		// Main start success, stop fallback.
		if (w.fallbackUpdater != nil) {
			w.fallbackUpdater.Stop()
		}
		return nil
	}
	fmt.Println("main start err", err)
	// Logger.e("Primary flag configs start failed, start fallback. Error: ", t)
	if (w.fallbackUpdater == nil) {
		// No fallback, main start failed is wrapper start fail
		return err
	}
	err = w.fallbackUpdater.Start(nil)
	if (err != nil) {
		return err
	}

	go func() {w.scheduleRetry()}()
	return nil
}

func (w *FlagConfigFallbackRetryWrapper) Stop() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if (w.retryTimer != nil) {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}
	w.mainUpdater.Stop()
	if (w.fallbackUpdater != nil) {
		w.fallbackUpdater.Stop()
	}
}

func (w *FlagConfigFallbackRetryWrapper) scheduleRetry() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if (w.retryTimer != nil) {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}
	w.retryTimer = time.AfterFunc(randTimeDuration(w.retryDelay, w.maxJitter), func() {
		fmt.Println("retrying")
		w.lock.Lock()
		defer w.lock.Unlock()

		if (w.retryTimer != nil) {
			w.retryTimer = nil
		}

		err := w.mainUpdater.Start(func (error) {
			go func() {w.scheduleRetry()}() // Don't care if poller start error or not, always retry.
			if (w.fallbackUpdater != nil) {
				w.fallbackUpdater.Start(nil)
			}
		})
		if (err == nil) {
			// Main start success, stop fallback.
			if (w.fallbackUpdater != nil) {
				w.fallbackUpdater.Stop()
			}
			return
		}
		
		go func() {w.scheduleRetry()}()
	})
}