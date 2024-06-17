package local

import (
	"sync"
)

// CohortStorage defines the interface for cohort storage operations
type CohortStorage interface {
	GetCohort(cohortID string) *Cohort
	GetCohorts() map[string]*Cohort
	GetCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{}
	GetCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{}
	PutCohort(cohort *Cohort)
	DeleteCohort(groupType, cohortID string)
}

// InMemoryCohortStorage is an in-memory implementation of CohortStorage
type InMemoryCohortStorage struct {
	lock               sync.RWMutex
	groupToCohortStore map[string]map[string]struct{}
	cohortStore        map[string]*Cohort
}

// NewInMemoryCohortStorage creates a new InMemoryCohortStorage instance
func NewInMemoryCohortStorage() *InMemoryCohortStorage {
	return &InMemoryCohortStorage{
		groupToCohortStore: make(map[string]map[string]struct{}),
		cohortStore:        make(map[string]*Cohort),
	}
}

// GetCohort retrieves a cohort by its ID
func (s *InMemoryCohortStorage) GetCohort(cohortID string) *Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cohortStore[cohortID]
}

// GetCohorts retrieves all cohorts
func (s *InMemoryCohortStorage) GetCohorts() map[string]*Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cohorts := make(map[string]*Cohort)
	for id, cohort := range s.cohortStore {
		cohorts[id] = cohort
	}
	return cohorts
}

// GetCohortsForUser retrieves cohorts for a user based on cohort IDs
func (s *InMemoryCohortStorage) GetCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{} {
	return s.GetCohortsForGroup(userGroupType, userID, cohortIDs)
}

// GetCohortsForGroup retrieves cohorts for a group based on cohort IDs
func (s *InMemoryCohortStorage) GetCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	s.lock.RLock()
	defer s.lock.RUnlock()
	groupTypeCohorts := s.groupToCohortStore[groupType]
	for cohortID := range groupTypeCohorts {
		if _, exists := cohortIDs[cohortID]; exists {
			if cohort, found := s.cohortStore[cohortID]; found {
				if _, memberExists := cohort.MemberIDs[groupName]; memberExists {
					result[cohortID] = struct{}{}
				}
			}
		}
	}
	return result
}

// PutCohort stores a cohort
func (s *InMemoryCohortStorage) PutCohort(cohort *Cohort) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, exists := s.groupToCohortStore[cohort.GroupType]; !exists {
		s.groupToCohortStore[cohort.GroupType] = make(map[string]struct{})
	}
	s.groupToCohortStore[cohort.GroupType][cohort.ID] = struct{}{}
	s.cohortStore[cohort.ID] = cohort
}

// DeleteCohort deletes a cohort by its ID and group type
func (s *InMemoryCohortStorage) DeleteCohort(groupType, cohortID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if groupCohorts, exists := s.groupToCohortStore[groupType]; exists {
		delete(groupCohorts, cohortID)
		if len(groupCohorts) == 0 {
			delete(s.groupToCohortStore, groupType)
		}
	}
	delete(s.cohortStore, cohortID)
}
