package local

import (
	"sync"
)

type CohortStorage interface {
	GetCohort(cohortID string) *Cohort
	GetCohorts() map[string]*Cohort
	GetCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{}
	GetCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{}
	PutCohort(cohort *Cohort)
	DeleteCohort(groupType, cohortID string)
}

type InMemoryCohortStorage struct {
	lock               sync.RWMutex
	groupToCohortStore map[string]map[string]struct{}
	cohortStore        map[string]*Cohort
}

func NewInMemoryCohortStorage() *InMemoryCohortStorage {
	return &InMemoryCohortStorage{
		groupToCohortStore: make(map[string]map[string]struct{}),
		cohortStore:        make(map[string]*Cohort),
	}
}

func (s *InMemoryCohortStorage) GetCohort(cohortID string) *Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cohortStore[cohortID]
}

func (s *InMemoryCohortStorage) GetCohorts() map[string]*Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cohorts := make(map[string]*Cohort)
	for id, cohort := range s.cohortStore {
		cohorts[id] = cohort
	}
	return cohorts
}

func (s *InMemoryCohortStorage) GetCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{} {
	return s.GetCohortsForGroup(userGroupType, userID, cohortIDs)
}

func (s *InMemoryCohortStorage) GetCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	s.lock.RLock()
	defer s.lock.RUnlock()

	groupTypeCohorts, groupExists := s.groupToCohortStore[groupType]
	if !groupExists {
		return result
	}

	for cohortID := range cohortIDs {
		if _, exists := groupTypeCohorts[cohortID]; exists {
			if cohort, found := s.cohortStore[cohortID]; found {
				for _, memberID := range cohort.MemberIds {
					if memberID == groupName {
						result[cohortID] = struct{}{}
						break
					}
				}
			}
		}
	}

	return result
}

func (s *InMemoryCohortStorage) PutCohort(cohort *Cohort) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, exists := s.groupToCohortStore[cohort.GroupType]; !exists {
		s.groupToCohortStore[cohort.GroupType] = make(map[string]struct{})
	}
	s.groupToCohortStore[cohort.GroupType][cohort.Id] = struct{}{}
	s.cohortStore[cohort.Id] = cohort
}

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
