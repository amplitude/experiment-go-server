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
	flagConfigStorage := NewInMemoryFlagConfigStorage()
	cohortStorage := NewInMemoryCohortStorage()
	cohortLoader := NewCohortLoader(cohortDownloadAPI, cohortStorage)

	runner := NewDeploymentRunner(
		&Config{},
		flagAPI,
		flagConfigStorage,
		cohortStorage,
		cohortLoader,
	)

	err := runner.Start()

	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestStartThrowsIfFirstCohortLoadFails(t *testing.T) {
	flagAPI := &mockFlagConfigApi{getFlagConfigsFunc: func() (map[string]*evaluation.Flag, error) {
		return map[string]*evaluation.Flag{"flag": createTestFlag()}, nil
	}}
	cohortDownloadAPI := &mockCohortDownloadApi{getCohortFunc: func(cohortID string, cohort *Cohort) (*Cohort, error) {
		return nil, errors.New("test")
	}}
	flagConfigStorage := NewInMemoryFlagConfigStorage()
	cohortStorage := NewInMemoryCohortStorage()
	cohortLoader := NewCohortLoader(cohortDownloadAPI, cohortStorage)

	runner := NewDeploymentRunner(
		&Config{},
		flagAPI,
		flagConfigStorage,
		cohortStorage,
		cohortLoader,
	)

	err := runner.Start()

	if err == nil {
		t.Error("Expected error but got nil")
	}
}

// Mock implementations for interfaces used in tests

type mockFlagConfigApi struct {
	getFlagConfigsFunc func() (map[string]*evaluation.Flag, error)
}

func (m *mockFlagConfigApi) GetFlagConfigs() (map[string]*evaluation.Flag, error) {
	if m.getFlagConfigsFunc != nil {
		return m.getFlagConfigsFunc()
	}
	return nil, fmt.Errorf("mock not implemented")
}

type mockCohortDownloadApi struct {
	getCohortFunc func(cohortID string, cohort *Cohort) (*Cohort, error)
}

func (m *mockCohortDownloadApi) GetCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
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
