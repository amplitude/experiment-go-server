package local

import (
	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/internal/logger"
	"sync"
)

type DeploymentRunner struct {
	config            *Config
	flagConfigApi     FlagConfigApi
	flagConfigStorage FlagConfigStorage
	cohortStorage     CohortStorage
	cohortLoader      *CohortLoader
	lock              sync.Mutex
	poller            *poller
	log               *logger.Log
}

func NewDeploymentRunner(
	config *Config,
	flagConfigApi FlagConfigApi,
	flagConfigStorage FlagConfigStorage,
	cohortStorage CohortStorage,
	cohortLoader *CohortLoader,
) *DeploymentRunner {
	dr := &DeploymentRunner{
		config:            config,
		flagConfigApi:     flagConfigApi,
		flagConfigStorage: flagConfigStorage,
		cohortStorage:     cohortStorage,
		cohortLoader:      cohortLoader,
		log:               logger.New(config.Debug),
	}
	dr.poller = newPoller()
	return dr
}

func (dr *DeploymentRunner) Start() error {
	dr.lock.Lock()
	defer dr.lock.Unlock()

	if err := dr.refresh(); err != nil {
		dr.log.Error("Initial refresh failed: %v", err)
		return err
	}

	dr.poller.Poll(dr.config.FlagConfigPollerInterval, func() {
		if err := dr.periodicRefresh(); err != nil {
			dr.log.Error("Periodic refresh failed: %v", err)
		}
	})
	return nil
}

func (dr *DeploymentRunner) periodicRefresh() error {
	defer func() {
		if r := recover(); r != nil {
			dr.log.Error("Recovered in periodicRefresh: %v", r)
		}
	}()
	return dr.refresh()
}

func (dr *DeploymentRunner) refresh() error {
	dr.log.Debug("Refreshing flag configs.")
	flagConfigs, err := dr.flagConfigApi.GetFlagConfigs()
	if err != nil {
		dr.log.Error("Failed to fetch flag configs: %v", err)
		return err
	}

	flagKeys := make(map[string]struct{})
	for _, flag := range flagConfigs {
		flagKeys[flag.Key] = struct{}{}
	}

	dr.flagConfigStorage.RemoveIf(func(f *evaluation.Flag) bool {
		_, exists := flagKeys[f.Key]
		return !exists
	})

	for _, flagConfig := range flagConfigs {
		cohortIDs := getAllCohortIDsFromFlag(flagConfig)
		if dr.cohortLoader == nil || len(cohortIDs) == 0 {
			dr.log.Debug("Putting non-cohort flag %s", flagConfig.Key)
			dr.flagConfigStorage.PutFlagConfig(flagConfig)
			continue
		}

		oldFlagConfig := dr.flagConfigStorage.GetFlagConfig(flagConfig.Key)

		err := dr.loadCohorts(*flagConfig, cohortIDs)
		if err != nil {
			dr.log.Error("Failed to load all cohorts for flag %s. Using the old flag config.", flagConfig.Key)
			dr.flagConfigStorage.PutFlagConfig(oldFlagConfig)
			return err
		}

		dr.flagConfigStorage.PutFlagConfig(flagConfig)
		dr.log.Debug("Stored flag config %s", flagConfig.Key)
	}

	dr.deleteUnusedCohorts()
	dr.log.Debug("Refreshed %d flag configs.", len(flagConfigs))
	return nil
}

func (dr *DeploymentRunner) loadCohorts(flagConfig evaluation.Flag, cohortIDs map[string]struct{}) error {
	task := func() error {
		for cohortID := range cohortIDs {
			task := dr.cohortLoader.LoadCohort(cohortID)
			err := task.Wait()
			if err != nil {
				if _, ok := err.(*CohortNotModifiedException); !ok {
					dr.log.Error("Failed to load cohort %s for flag %s: %v", cohortID, flagConfig.Key, err)
					return err
				}
				continue
			}
			dr.log.Debug("Cohort %s loaded for flag %s", cohortID, flagConfig.Key)
		}
		return nil
	}

	// Using a goroutine to simulate async task execution
	errCh := make(chan error)
	go func() {
		errCh <- task()
	}()
	err := <-errCh
	return err
}

func (dr *DeploymentRunner) deleteUnusedCohorts() {
	flagCohortIDs := make(map[string]struct{})
	for _, flag := range dr.flagConfigStorage.GetFlagConfigs() {
		for cohortID := range getAllCohortIDsFromFlag(flag) {
			flagCohortIDs[cohortID] = struct{}{}
		}
	}

	storageCohorts := dr.cohortStorage.GetCohorts()
	for cohortID := range storageCohorts {
		if _, exists := flagCohortIDs[cohortID]; !exists {
			cohort := storageCohorts[cohortID]
			if cohort != nil {
				dr.cohortStorage.DeleteCohort(cohort.GroupType, cohortID)
			}
		}
	}
}
