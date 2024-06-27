package local

import (
	"sync"
	"sync/atomic"
)

type CohortLoader struct {
	cohortDownloadApi CohortDownloadApi
	cohortStorage     CohortStorage
	jobs              sync.Map
	executor          *sync.Pool
	lockJobs          sync.Mutex
}

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

func (cl *CohortLoader) removeJob(cohortID string) {
	cl.jobs.Delete(cohortID)
}

type CohortLoaderTask struct {
	loader   *CohortLoader
	cohortID string
	done     int32
	doneChan chan struct{}
	err      error
}

func (task *CohortLoaderTask) init(loader *CohortLoader, cohortID string) {
	task.loader = loader
	task.cohortID = cohortID
	task.done = 0
	task.doneChan = make(chan struct{})
	task.err = nil
}

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

func (task *CohortLoaderTask) Wait() error {
	<-task.doneChan
	return task.err
}

func (cl *CohortLoader) downloadCohort(cohortID string) (*Cohort, error) {
	cohort := cl.cohortStorage.GetCohort(cohortID)
	return cl.cohortDownloadApi.GetCohort(cohortID, cohort)
}
