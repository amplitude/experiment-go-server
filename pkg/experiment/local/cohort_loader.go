package local

import (
	"sync"
	"sync/atomic"
)

// CohortLoader handles the loading of cohorts using CohortDownloadApi and CohortStorage.
type CohortLoader struct {
	cohortDownloadApi CohortDownloadApi
	cohortStorage     CohortStorage
	jobs              sync.Map
	executor          *sync.Pool
	lockJobs          sync.Mutex
}

// NewCohortLoader creates a new instance of CohortLoader.
func NewCohortLoader(cohortDownloadApi CohortDownloadApi, cohortStorage CohortStorage) *CohortLoader {
	return &CohortLoader{
		cohortDownloadApi: cohortDownloadApi,
		cohortStorage:     cohortStorage,
		executor: &sync.Pool{
			New: func() interface{} {
				return &CohortLoaderTask{}
			},
		},
	}
}

// LoadCohort initiates the loading of a cohort.
func (cl *CohortLoader) LoadCohort(cohortID string) *CohortLoaderTask {
	cl.lockJobs.Lock()
	defer cl.lockJobs.Unlock()

	task, ok := cl.jobs.Load(cohortID)
	if !ok {
		task = cl.executor.Get().(*CohortLoaderTask)
		task.(*CohortLoaderTask).init(cl, cohortID)
		cl.jobs.Store(cohortID, task)
		go task.(*CohortLoaderTask).run()
	}

	return task.(*CohortLoaderTask)
}

// removeJob removes a job from the jobs map.
func (cl *CohortLoader) removeJob(cohortID string) {
	cl.jobs.Delete(cohortID)
}

// CohortLoaderTask represents a task for loading a cohort.
type CohortLoaderTask struct {
	loader   *CohortLoader
	cohortID string
	done     int32
	doneChan chan struct{}
	err      error
}

// init initializes a CohortLoaderTask.
func (task *CohortLoaderTask) init(loader *CohortLoader, cohortID string) {
	task.loader = loader
	task.cohortID = cohortID
	task.done = 0
	task.doneChan = make(chan struct{})
	task.err = nil
}

// run executes the task of loading a cohort.
func (task *CohortLoaderTask) run() {
	defer task.loader.executor.Put(task)

	cohort, err := task.loader.downloadCohort(task.cohortID)
	if err != nil {
		task.err = err
	} else {
		task.loader.cohortStorage.PutCohort(cohort)
	}

	task.loader.removeJob(task.cohortID)
	atomic.StoreInt32(&task.done, 1)
	close(task.doneChan)
}

// Wait waits for the task to complete.
func (task *CohortLoaderTask) Wait() error {
	<-task.doneChan
	return task.err
}

// downloadCohort downloads a cohort.
func (cl *CohortLoader) downloadCohort(cohortID string) (*Cohort, error) {
	cohort := cl.cohortStorage.GetCohort(cohortID)
	return cl.cohortDownloadApi.GetCohort(cohortID, cohort)
}

type CohortDownloadApiImpl struct{}

// GetCohort gets a cohort.
func (api *CohortDownloadApiImpl) GetCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	// Placeholder implementation
	return &Cohort{
		ID:           cohortID,
		LastModified: 0,
		Size:         0,
		MemberIDs:    make(map[string]struct{}),
		GroupType:    "example",
	}, nil
}
