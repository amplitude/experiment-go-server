package local

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type cohortLoader struct {
	cohortDownloadApi cohortDownloadApi
	cohortStorage     CohortStorage
	jobs              sync.Map
	executor          *sync.Pool
	lockJobs          sync.Mutex
}

func newCohortLoader(cohortDownloadApi cohortDownloadApi, cohortStorage CohortStorage) *cohortLoader {
	return &cohortLoader{
		cohortDownloadApi: cohortDownloadApi,
		cohortStorage:     cohortStorage,
		executor: &sync.Pool{
			New: func() interface{} {
				return &CohortLoaderTask{}
			},
		},
	}
}

func (cl *cohortLoader) loadCohort(cohortId string) *CohortLoaderTask {
	cl.lockJobs.Lock()
	defer cl.lockJobs.Unlock()

	task, ok := cl.jobs.Load(cohortId)
	if !ok {
		task = cl.executor.Get().(*CohortLoaderTask)
		task.(*CohortLoaderTask).init(cl, cohortId)
		cl.jobs.Store(cohortId, task)
		go task.(*CohortLoaderTask).run()
	}

	return task.(*CohortLoaderTask)
}

func (cl *cohortLoader) removeJob(cohortId string) {
	cl.jobs.Delete(cohortId)
}

type CohortLoaderTask struct {
	loader   *cohortLoader
	cohortId string
	done     int32
	doneChan chan struct{}
	err      error
}

func (task *CohortLoaderTask) init(loader *cohortLoader, cohortId string) {
	task.loader = loader
	task.cohortId = cohortId
	task.done = 0
	task.doneChan = make(chan struct{})
	task.err = nil
}

func (task *CohortLoaderTask) run() {
	defer task.loader.executor.Put(task)

	cohort, err := task.loader.downloadCohort(task.cohortId)
	if err != nil {
		task.err = err
	} else {
		if cohort != nil {
			task.loader.cohortStorage.putCohort(cohort)
		}
	}

	task.loader.removeJob(task.cohortId)
	atomic.StoreInt32(&task.done, 1)
	close(task.doneChan)
}

func (task *CohortLoaderTask) wait() error {
	<-task.doneChan
	return task.err
}

func (cl *cohortLoader) downloadCohort(cohortID string) (*Cohort, error) {
	cohort := cl.cohortStorage.getCohort(cohortID)
	return cl.cohortDownloadApi.getCohort(cohortID, cohort)
}

func (cl *cohortLoader) updateStoredCohorts() error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(cl.cohortStorage.getCohortIds()))

	cohortIds := make([]string, 0, len(cl.cohortStorage.getCohortIds()))
	for id := range cl.cohortStorage.getCohortIds() {
		cohortIds = append(cohortIds, id)
	}

	for _, cohortID := range cohortIds {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			task := cl.loadCohort(id)
			if err := task.wait(); err != nil {
				errorChan <- fmt.Errorf("cohort %s: %v", id, err)
			}
		}(cohortID)
	}

	wg.Wait()
	close(errorChan)

	var errorMessages []string
	for err := range errorChan {
		errorMessages = append(errorMessages, err.Error())
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("One or more cohorts failed to download:\n%s",
			strings.Join(errorMessages, "\n"))
	}
	return nil
}
