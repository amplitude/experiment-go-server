package local

import (
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
	Start(func(error)) error
	Stop()
}

// The base for all flag config updaters.
// Contains a method to properly update the flag configs into storage and download cohorts.
type flagConfigUpdaterBase struct {
	flagConfigStorage flagConfigStorage
	cohortStorage     cohortStorage
	cohortLoader      *cohortLoader
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
		cohortStorage:     cohortStorage,
		cohortLoader:      cohortLoader,
		log:               logger.New(config.Debug),
	}
}

// Updates the received flag configs into storage and download cohorts.
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

// The streamer for flag configs. It receives flag configs through server side events.
type flagConfigStreamer struct {
	flagConfigUpdaterBase
	flagConfigStreamApi flagConfigStreamApi
	lock                sync.Mutex
}

func newFlagConfigStreamer(
	flagConfigStreamApi flagConfigStreamApi,
	config *Config,
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
) flagConfigUpdater {
	return &flagConfigStreamer{
		flagConfigStreamApi:   flagConfigStreamApi,
		flagConfigUpdaterBase: newFlagConfigUpdaterBase(flagConfigStorage, cohortStorage, cohortLoader, config),
	}
}

func (s *flagConfigStreamer) Start(onError func(error)) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stopInternal()
	return s.flagConfigStreamApi.Connect(
		func(flags map[string]*evaluation.Flag) error {
			return s.update(flags)
		},
		func(flags map[string]*evaluation.Flag) error {
			return s.update(flags)
		},
		func(err error) {
			s.Stop()
			if onError != nil {
				go func() {onError(err)}()
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

// The poller for flag configs. It polls every configured interval.
// On start, it polls a set of flag configs. If failed, error is returned. If success, poller starts.
type flagConfigPoller struct {
	flagConfigUpdaterBase
	flagConfigApi flagConfigApi
	config        *Config
	poller        *poller
	lock          sync.Mutex
}

func newFlagConfigPoller(
	flagConfigApi flagConfigApi,
	config *Config,
	flagConfigStorage flagConfigStorage,
	cohortStorage cohortStorage,
	cohortLoader *cohortLoader,
) flagConfigUpdater {
	return &flagConfigPoller{
		flagConfigApi:         flagConfigApi,
		config:                config,
		flagConfigUpdaterBase: newFlagConfigUpdaterBase(flagConfigStorage, cohortStorage, cohortLoader, config),
	}
}

func (p *flagConfigPoller) Start(onError func(error)) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.stopInternal()

	if err := p.updateFlagConfigs(); err != nil {
		p.log.Error("Initial updateFlagConfigs failed: %v", err)
		return err
	}

	p.poller = newPoller()
	p.poller.Poll(p.config.FlagConfigPollerInterval, func() {
		if err := p.periodicRefresh(); err != nil {
			p.log.Error("Periodic updateFlagConfigs failed: %v", err)
			p.Stop()
			if (onError != nil) {
				go func() {onError(err)}()
			}
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

func (p *flagConfigPoller) stopInternal() {
	if p.poller != nil {
		close(p.poller.shutdown)
		p.poller = nil
	}
}

func (p *flagConfigPoller) Stop() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.stopInternal()
}

// A wrapper around flag config updaters to retry and fallback.
// If the main updater fails, it will fallback to the fallback updater and main updater enters retry loop.
type flagConfigFallbackRetryWrapper struct {
	log             *logger.Log
	mainUpdater     flagConfigUpdater
	fallbackUpdater flagConfigUpdater
	retryDelay      time.Duration
	maxJitter       time.Duration
	retryTimer      *time.Timer
	fallbackStartRetryDelay      time.Duration
	fallbackStartRetryMaxJitter       time.Duration
	fallbackStartRetryTimer *time.Timer
	lock            sync.Mutex
	isRunning       bool
}

func newflagConfigFallbackRetryWrapper(
	mainUpdater flagConfigUpdater,
	fallbackUpdater flagConfigUpdater,
	retryDelay time.Duration,
	maxJitter time.Duration,
	fallbackStartRetryDelay      time.Duration,
	fallbackStartRetryMaxJitter       time.Duration,
	debug bool,
) flagConfigUpdater {
	return &flagConfigFallbackRetryWrapper{
		log:             logger.New(debug),
		mainUpdater:     mainUpdater,
		fallbackUpdater: fallbackUpdater,
		retryDelay:      retryDelay,
		maxJitter:       maxJitter,
		fallbackStartRetryDelay:      fallbackStartRetryDelay,
		fallbackStartRetryMaxJitter:       fallbackStartRetryMaxJitter,
		isRunning: false,
	}
}

/**
 * Start tries to start main updater first.
 * If it failed, start the fallback updater.
 *   If fallback updater failed as well, return error.
 *   If fallback updater succeed, main updater enters retry, return ok.
 * After started, if main failed, main enters retry loop and fallback will start.
 *   If fallback start failed, fallback will enter start retry loop until it's successfully started.
 *   If fallback start success, but failed later, it's not monitored. It's recommended to wrap fallback with flagConfigFallbackRetryWrapper.
 * Since the wrapper retries, so there will never be error case.
 * Thus, onError will never be called.
 */
func (w *flagConfigFallbackRetryWrapper) Start(onError func(error)) error {
	// if (mainUpdater is flagConfigFallbackRetryWrapper) {
	//     return errors.New("Do not use flagConfigFallbackRetryWrapper as main updater. Fallback updater will never be used. Rewrite retry and fallback logic.")
	// }

	w.lock.Lock()
	defer w.lock.Unlock()

	if w.retryTimer != nil {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}

	err := w.mainUpdater.Start(func(err error) {
		w.log.Debug("main updater updating err, starting fallback if available. error: ", err)
		go func() { w.scheduleRetry() }() // Don't care if poller start error or not, always retry.
		go func() { w.fallbackStart() }()
	})
	if err == nil {
		// Main start success, stop fallback.
		if w.fallbackStartRetryTimer != nil {
			w.fallbackStartRetryTimer.Stop()
		}
		if w.fallbackUpdater != nil {
			w.fallbackUpdater.Stop()
		}
		w.isRunning = true
		return nil
	}
	if w.fallbackUpdater == nil {
		// No fallback, main start failed is wrapper start fail
		w.log.Error("main updater start err, no fallback. error: ", err)
		return err
	}
	w.log.Debug("main updater start err, starting fallback. error: ", err)
	err = w.fallbackUpdater.Start(nil)
	if err != nil {
		w.log.Debug("fallback updater start failed. error: ", err)
		return err
	}

	w.isRunning = true
	go func() { w.scheduleRetry() }()
	return nil
}

func (w *flagConfigFallbackRetryWrapper) Stop() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.isRunning = false

	if w.retryTimer != nil {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}
	w.mainUpdater.Stop()
	if w.fallbackStartRetryTimer != nil {
		w.fallbackStartRetryTimer.Stop()
	}
	if w.fallbackUpdater != nil {
		w.fallbackUpdater.Stop()
	}
}

func (w *flagConfigFallbackRetryWrapper) scheduleRetry() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if (!w.isRunning) {
		return
	}

	if w.retryTimer != nil {
		w.retryTimer.Stop()
		w.retryTimer = nil
	}
	w.retryTimer = time.AfterFunc(randTimeDuration(w.retryDelay, w.maxJitter), func() {
		w.lock.Lock()
		defer w.lock.Unlock()

		if (!w.isRunning) {
			return
		}

		if w.retryTimer != nil {
			w.retryTimer = nil
		}

		w.log.Debug("main updater retry start")
		err := w.mainUpdater.Start(func(err error) {
			w.log.Debug("main updater updating err, starting fallback if available. error: ", err)
			go func() { w.scheduleRetry() }() // Don't care if poller start error or not, always retry.
			go func() { w.fallbackStart() }()
		})
		if err == nil {
			// Main start success, stop fallback.
			w.log.Debug("main updater retry start success")
			if w.fallbackStartRetryTimer != nil {
				w.fallbackStartRetryTimer.Stop()
			}
			if w.fallbackUpdater != nil {
				w.fallbackUpdater.Stop()
			}
			return
		}

		go func() { w.scheduleRetry() }()
	})
}

func (w *flagConfigFallbackRetryWrapper) fallbackStart() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if (!w.isRunning) {
		return
	}
	if (w.fallbackUpdater == nil) {
		return
	}

	err := w.fallbackUpdater.Start(nil)
	if (err != nil) {
		w.log.Debug("fallback updater start failed and scheduling retry")
		w.fallbackStartRetryTimer = time.AfterFunc(randTimeDuration(w.fallbackStartRetryDelay, w.fallbackStartRetryMaxJitter), func() {
			w.fallbackStart()
		})
	}
}
