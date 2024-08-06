package local

import (
	"fmt"
	"strings"
	"sync"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/internal/logger"
)

type deploymentRunner struct {
	config            *Config
	flagConfigApi     flagConfigApi
	flagConfigStorage flagConfigStorage
	cohortStorage     CohortStorage
	cohortLoader      *cohortLoader
	lock              sync.Mutex
	poller            *poller
	log               *logger.Log
}

func newDeploymentRunner(
	config *Config,
	flagConfigApi flagConfigApi,
	flagConfigStorage flagConfigStorage,
	cohortStorage CohortStorage,
	cohortLoader *cohortLoader,
) *deploymentRunner {
	dr := &deploymentRunner{
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

func (dr *deploymentRunner) start() error {
	dr.lock.Lock()
	defer dr.lock.Unlock()

	if err := dr.updateFlagConfigs(); err != nil {
		dr.log.Error("Initial updateFlagConfigs failed: %v", err)
		return err
	}

	dr.poller.Poll(dr.config.FlagConfigPollerInterval, func() {
		if err := dr.periodicRefresh(); err != nil {
			dr.log.Error("Periodic updateFlagConfigs failed: %v", err)
		}
	})

	if dr.cohortLoader != nil {
		dr.poller.Poll(dr.config.FlagConfigPollerInterval, func() {
			dr.updateStoredCohorts()
		})
	}
	return nil
}

func (dr *deploymentRunner) periodicRefresh() error {
	defer func() {
		if r := recover(); r != nil {
			dr.log.Error("Recovered in periodicRefresh: %v", r)
		}
	}()
	return dr.updateFlagConfigs()
}

func (dr *deploymentRunner) updateFlagConfigs() error {
	dr.log.Debug("Refreshing flag configs.")
	flagConfigs, err := dr.flagConfigApi.getFlagConfigs()
	if err != nil {
		dr.log.Error("Failed to fetch flag configs: %v", err)
		return err
	}

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
	dr.downloadCohorts(cohortIDsToDownload)

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

func (dr *deploymentRunner) updateStoredCohorts() {
	cohortIDs := getAllCohortIDsFromFlags(dr.flagConfigStorage.getFlagConfigsArray())
	dr.downloadCohorts(cohortIDs)
}

func (dr *deploymentRunner) deleteUnusedCohorts() {
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

func difference(set1, set2 map[string]struct{}) map[string]struct{} {
	diff := make(map[string]struct{})
	for k := range set1 {
		if _, exists := set2[k]; !exists {
			diff[k] = struct{}{}
		}
	}
	return diff
}

func (dr *deploymentRunner) downloadCohorts(cohortIDs map[string]struct{}) {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(cohortIDs))

	for cohortID := range cohortIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			task := dr.cohortLoader.loadCohort(id)
			if err := task.wait(); err != nil {
				errorChan <- fmt.Errorf("cohort %s: %v", id, err)
			}
		}(cohortID)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	var errorMessages []string
	for err := range errorChan {
		errorMessages = append(errorMessages, err.Error())
		dr.log.Error("Error downloading cohort: %v", err)
	}

	if len(errorMessages) > 0 {
		dr.log.Error("One or more cohorts failed to download:\n%s", strings.Join(errorMessages, "\n"))
	}
}
