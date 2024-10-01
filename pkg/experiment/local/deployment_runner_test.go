package local

import (
	"errors"
	"fmt"
	"testing"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

const (
	CohortId = "1234"
)

func TestStartThrowsIfFirstFlagConfigLoadFails(t *testing.T) {
	flagAPI := &mockFlagConfigApi{getFlagConfigsFunc: func() (map[string]*evaluation.Flag, error) {
		return nil, errors.New("test")
	}}
	cohortDownloadAPI := &mockCohortDownloadApi{}
	flagConfigStorage := newInMemoryFlagConfigStorage()
	cohortStorage := newInMemoryCohortStorage()
	cohortLoader := newCohortLoader(cohortDownloadAPI, cohortStorage)

	runner := newDeploymentRunner(
		&Config{},
		flagAPI,
		nil,
		flagConfigStorage,
		cohortStorage,
		cohortLoader,
	)

	err := runner.start()

	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestStartSucceedsEvenIfFirstCohortLoadFails(t *testing.T) {
	flagAPI := &mockFlagConfigApi{getFlagConfigsFunc: func() (map[string]*evaluation.Flag, error) {
		return map[string]*evaluation.Flag{"flag": createTestFlag()}, nil
	}}
	cohortDownloadAPI := &mockCohortDownloadApi{getCohortFunc: func(cohortID string, cohort *Cohort) (*Cohort, error) {
		return nil, errors.New("test")
	}}
	flagConfigStorage := newInMemoryFlagConfigStorage()
	cohortStorage := newInMemoryCohortStorage()
	cohortLoader := newCohortLoader(cohortDownloadAPI, cohortStorage)

	runner := newDeploymentRunner(
		DefaultConfig,
		flagAPI,
		nil,
		flagConfigStorage,
		cohortStorage,
		cohortLoader,
	)

	err := runner.start()

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}
}

type mockFlagConfigApi struct {
	getFlagConfigsFunc func() (map[string]*evaluation.Flag, error)
}

func (m *mockFlagConfigApi) getFlagConfigs() (map[string]*evaluation.Flag, error) {
	if m.getFlagConfigsFunc != nil {
		return m.getFlagConfigsFunc()
	}
	return nil, fmt.Errorf("mock not implemented")
}

type mockCohortDownloadApi struct {
	getCohortFunc func(cohortID string, cohort *Cohort) (*Cohort, error)
}

func (m *mockCohortDownloadApi) getCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	if m.getCohortFunc != nil {
		return m.getCohortFunc(cohortID, cohort)
	}
	return nil, fmt.Errorf("mock not implemented")
}

func createTestFlag() *evaluation.Flag {
	return &evaluation.Flag{
		Key:      "flag",
		Variants: map[string]*evaluation.Variant{},
		Segments: []*evaluation.Segment{
			{
				Conditions: [][]*evaluation.Condition{
					{
						{
							Selector: []string{"context", "user", "cohort_ids"},
							Op:       "set contains any",
							Values:   []string{CohortId},
						},
					},
				},
			},
		},
	}
}
