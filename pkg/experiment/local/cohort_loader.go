package local

import (
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
