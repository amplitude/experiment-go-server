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
	var cohortDownloadErrors []string

	// Download all new cohorts
	for cohortID := range cohortIDsToDownload {
		if err := dr.cohortLoader.loadCohort(cohortID).wait(); err != nil {
			cohortDownloadErrors = append(cohortDownloadErrors, fmt.Sprintf("Cohort %s: %v", cohortID, err))
			dr.log.Error("Download cohort %s failed: %v", cohortID, err)
		}
	}

	// Get updated set of cohort ids
	updatedCohortIDs := dr.cohortStorage.getCohortIds()
	// Iterate through new flag configs and check if their required cohorts exist
	failedFlagCount := 0
	for _, flagConfig := range flagConfigs {
		cohortIDs := getAllCohortIDsFromFlag(flagConfig)
		if len(cohortIDs) == 0 || dr.cohortLoader == nil {
			dr.flagConfigStorage.putFlagConfig(flagConfig)
			dr.log.Debug("Putting non-cohort flag %s", flagConfig.Key)
		} else if subset(cohortIDs, updatedCohortIDs) {
			dr.flagConfigStorage.putFlagConfig(flagConfig)
			dr.log.Debug("Putting flag %s", flagConfig.Key)
		} else {
			dr.log.Error("Flag %s not updated because not all required cohorts could be loaded", flagConfig.Key)
			failedFlagCount++
		}
	}

	// Delete unused cohorts
	dr.deleteUnusedCohorts()
	dr.log.Debug("Refreshed %d flag configs.", len(flagConfigs)-failedFlagCount)

	// If there are any download errors, raise an aggregated exception
	if len(cohortDownloadErrors) > 0 {
		errorCount := len(cohortDownloadErrors)
		errorMessages := strings.Join(cohortDownloadErrors, "\n")
		return fmt.Errorf("%d cohort(s) failed to download:\n%s", errorCount, errorMessages)
	}

	return nil
}

func (dr *deploymentRunner) updateStoredCohorts() {
	err := dr.cohortLoader.updateStoredCohorts()
	if err != nil {
		dr.log.Error("Error updating stored cohorts: %v", err)
	}
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

func subset(subset, set map[string]struct{}) bool {
	for k := range subset {
		if _, exists := set[k]; !exists {
			return false
		}
	}
	return true
}
