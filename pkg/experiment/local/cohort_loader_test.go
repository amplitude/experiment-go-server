package local

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestLoadSuccess(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := newInMemoryCohortStorage()
	loader := newCohortLoader(api, storage)

	// Define mock behavior
	api.On("getCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)
	api.On("getCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "b", LastModified: 0, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}, nil)

	futureA := loader.loadCohort("a")
	futureB := loader.loadCohort("b")

	if err := futureA.wait(); err != nil {
		t.Errorf("futureA.wait() returned error: %v", err)
	}
	if err := futureB.wait(); err != nil {
		t.Errorf("futureB.wait() returned error: %v", err)
	}

	storageDescriptionA := storage.getCohort("a")
	storageDescriptionB := storage.getCohort("b")
	expectedA := &Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}
	expectedB := &Cohort{Id: "b", LastModified: 0, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}

	if !CohortEquals(storageDescriptionA, expectedA) {
		t.Errorf("Unexpected cohort A stored: %+v", storageDescriptionA)
	}
	if !CohortEquals(storageDescriptionB, expectedB) {
		t.Errorf("Unexpected cohort B stored: %+v", storageDescriptionB)
	}

	storageUser1Cohorts := storage.getCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}})
	storageUser2Cohorts := storage.getCohortsForUser("2", map[string]struct{}{"a": {}, "b": {}})
	if len(storageUser1Cohorts) != 2 || len(storageUser2Cohorts) != 1 {
		t.Errorf("Unexpected user cohorts: User1: %+v, User2: %+v", storageUser1Cohorts, storageUser2Cohorts)
	}
}

func TestFilterCohortsAlreadyComputed(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := newInMemoryCohortStorage()
	loader := newCohortLoader(api, storage)

	storage.putCohort(&Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}})
	storage.putCohort(&Cohort{Id: "b", LastModified: 0, Size: 0, MemberIds: []string{}})

	// Define mock behavior
	api.On("getCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}, GroupType: userGroupType}, nil)
	api.On("getCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "b", LastModified: 1, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}, nil)

	futureA := loader.loadCohort("a")
	futureB := loader.loadCohort("b")

	if err := futureA.wait(); err != nil {
		t.Errorf("futureA.wait() returned error: %v", err)
	}
	if err := futureB.wait(); err != nil {
		t.Errorf("futureB.wait() returned error: %v", err)
	}

	storageDescriptionA := storage.getCohort("a")
	storageDescriptionB := storage.getCohort("b")
	expectedA := &Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}, GroupType: userGroupType}
	expectedB := &Cohort{Id: "b", LastModified: 1, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}

	if !CohortEquals(storageDescriptionA, expectedA) {
		t.Errorf("Unexpected cohort A stored: %+v", storageDescriptionA)
	}
	if !CohortEquals(storageDescriptionB, expectedB) {
		t.Errorf("Unexpected cohort B stored: %+v", storageDescriptionB)
	}

	storageUser1Cohorts := storage.getCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}})
	storageUser2Cohorts := storage.getCohortsForUser("2", map[string]struct{}{"a": {}, "b": {}})
	if len(storageUser1Cohorts) != 1 || len(storageUser2Cohorts) != 1 {
		t.Errorf("Unexpected user cohorts: User1: %+v, User2: %+v", storageUser1Cohorts, storageUser2Cohorts)
	}
}

func TestLoadDownloadFailureThrows(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := newInMemoryCohortStorage()
	loader := newCohortLoader(api, storage)

	// Define mock behavior
	api.On("getCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)
	api.On("getCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(nil, errors.New("connection timed out"))
	api.On("getCohort", "c", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "c", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)

	futureA := loader.loadCohort("a")
	errB := loader.loadCohort("b").wait()
	futureC := loader.loadCohort("c")

	if err := futureA.wait(); err != nil {
		t.Errorf("futureA.wait() returned error: %v", err)
	}

	if errB == nil || errB.Error() != "connection timed out" {
		t.Errorf("futureB.wait() expected 'Connection timed out' error, got: %v", errB)
	}

	if err := futureC.wait(); err != nil {
		t.Errorf("futureC.wait() returned error: %v", err)
	}

	expectedCohorts := map[string]struct{}{"a": {}, "c": {}}
	actualCohorts := storage.getCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}, "c": {}})
	if len(actualCohorts) != len(expectedCohorts) {
		t.Errorf("Expected cohorts for user '1': %+v, but got: %+v", expectedCohorts, actualCohorts)
	}
}
