package local

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestLoadSuccess(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := NewInMemoryCohortStorage()
	loader := NewCohortLoader(api, storage)

	// Define mock behavior
	api.On("GetCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)
	api.On("GetCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "b", LastModified: 0, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}, nil)

	futureA := loader.LoadCohort("a")
	futureB := loader.LoadCohort("b")

	if err := futureA.Wait(); err != nil {
		t.Errorf("futureA.Wait() returned error: %v", err)
	}
	if err := futureB.Wait(); err != nil {
		t.Errorf("futureB.Wait() returned error: %v", err)
	}

	storageDescriptionA := storage.GetCohort("a")
	storageDescriptionB := storage.GetCohort("b")
	expectedA := &Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}
	expectedB := &Cohort{Id: "b", LastModified: 0, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}

	if !CohortEquals(storageDescriptionA, expectedA) {
		t.Errorf("Unexpected cohort A stored: %+v", storageDescriptionA)
	}
	if !CohortEquals(storageDescriptionB, expectedB) {
		t.Errorf("Unexpected cohort B stored: %+v", storageDescriptionB)
	}

	storageUser1Cohorts := storage.GetCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}})
	storageUser2Cohorts := storage.GetCohortsForUser("2", map[string]struct{}{"a": {}, "b": {}})
	if len(storageUser1Cohorts) != 2 || len(storageUser2Cohorts) != 1 {
		t.Errorf("Unexpected user cohorts: User1: %+v, User2: %+v", storageUser1Cohorts, storageUser2Cohorts)
	}
}

func TestFilterCohortsAlreadyComputed(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := NewInMemoryCohortStorage()
	loader := NewCohortLoader(api, storage)

	storage.PutCohort(&Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}})
	storage.PutCohort(&Cohort{Id: "b", LastModified: 0, Size: 0, MemberIds: []string{}})

	// Define mock behavior
	api.On("GetCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}, GroupType: userGroupType}, nil)
	api.On("GetCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "b", LastModified: 1, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}, nil)

	futureA := loader.LoadCohort("a")
	futureB := loader.LoadCohort("b")

	if err := futureA.Wait(); err != nil {
		t.Errorf("futureA.Wait() returned error: %v", err)
	}
	if err := futureB.Wait(); err != nil {
		t.Errorf("futureB.Wait() returned error: %v", err)
	}

	storageDescriptionA := storage.GetCohort("a")
	storageDescriptionB := storage.GetCohort("b")
	expectedA := &Cohort{Id: "a", LastModified: 0, Size: 0, MemberIds: []string{}, GroupType: userGroupType}
	expectedB := &Cohort{Id: "b", LastModified: 1, Size: 2, MemberIds: []string{"1", "2"}, GroupType: userGroupType}

	if !CohortEquals(storageDescriptionA, expectedA) {
		t.Errorf("Unexpected cohort A stored: %+v", storageDescriptionA)
	}
	if !CohortEquals(storageDescriptionB, expectedB) {
		t.Errorf("Unexpected cohort B stored: %+v", storageDescriptionB)
	}

	storageUser1Cohorts := storage.GetCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}})
	storageUser2Cohorts := storage.GetCohortsForUser("2", map[string]struct{}{"a": {}, "b": {}})
	if len(storageUser1Cohorts) != 1 || len(storageUser2Cohorts) != 1 {
		t.Errorf("Unexpected user cohorts: User1: %+v, User2: %+v", storageUser1Cohorts, storageUser2Cohorts)
	}
}

func TestLoadDownloadFailureThrows(t *testing.T) {
	api := &MockCohortDownloadApi{}
	storage := NewInMemoryCohortStorage()
	loader := NewCohortLoader(api, storage)

	// Define mock behavior
	api.On("GetCohort", "a", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "a", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)
	api.On("GetCohort", "b", mock.AnythingOfType("*local.Cohort")).Return(nil, errors.New("connection timed out"))
	api.On("GetCohort", "c", mock.AnythingOfType("*local.Cohort")).Return(&Cohort{Id: "c", LastModified: 0, Size: 1, MemberIds: []string{"1"}, GroupType: userGroupType}, nil)

	futureA := loader.LoadCohort("a")
	errB := loader.LoadCohort("b").Wait()
	futureC := loader.LoadCohort("c")

	if err := futureA.Wait(); err != nil {
		t.Errorf("futureA.Wait() returned error: %v", err)
	}

	if errB == nil || errB.Error() != "connection timed out" {
		t.Errorf("futureB.Wait() expected 'Connection timed out' error, got: %v", errB)
	}

	if err := futureC.Wait(); err != nil {
		t.Errorf("futureC.Wait() returned error: %v", err)
	}

	expectedCohorts := map[string]struct{}{"a": {}, "c": {}}
	actualCohorts := storage.GetCohortsForUser("1", map[string]struct{}{"a": {}, "b": {}, "c": {}})
	if len(actualCohorts) != len(expectedCohorts) {
		t.Errorf("Expected cohorts for user '1': %+v, but got: %+v", expectedCohorts, actualCohorts)
	}
}
