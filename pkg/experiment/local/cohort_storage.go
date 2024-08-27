package local

import (
	"sync"
)

type cohortStorage interface {
	getCohort(cohortID string) *Cohort
	getCohorts() map[string]*Cohort
	getCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{}
	getCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{}
	putCohort(cohort *Cohort)
	deleteCohort(groupType, cohortID string)
	getCohortIds() map[string]struct{}
}

type inMemoryCohortStorage struct {
	lock               sync.RWMutex
	groupToCohortStore map[string]map[string]struct{}
	cohortStore        map[string]*Cohort
}

func newInMemoryCohortStorage() *inMemoryCohortStorage {
	return &inMemoryCohortStorage{
		groupToCohortStore: make(map[string]map[string]struct{}),
		cohortStore:        make(map[string]*Cohort),
	}
}

func (s *inMemoryCohortStorage) getCohort(cohortID string) *Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.cohortStore[cohortID]
}

func (s *inMemoryCohortStorage) getCohorts() map[string]*Cohort {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cohorts := make(map[string]*Cohort)
	for id, cohort := range s.cohortStore {
		cohorts[id] = cohort
	}
	return cohorts
}

func (s *inMemoryCohortStorage) getCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{} {
	return s.getCohortsForGroup(userGroupType, userID, cohortIDs)
}

func (s *inMemoryCohortStorage) getCohortsForGroup(groupType, groupName string, cohortIDs map[string]struct{}) map[string]struct{} {
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

func (s *inMemoryCohortStorage) putCohort(cohort *Cohort) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, exists := s.groupToCohortStore[cohort.GroupType]; !exists {
		s.groupToCohortStore[cohort.GroupType] = make(map[string]struct{})
	}
	s.groupToCohortStore[cohort.GroupType][cohort.Id] = struct{}{}
	s.cohortStore[cohort.Id] = cohort
}

func (s *inMemoryCohortStorage) deleteCohort(groupType, cohortID string) {
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

func (s *inMemoryCohortStorage) getCohortIds() map[string]struct{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cohortIds := make(map[string]struct{})
	for id := range s.cohortStore {
		cohortIds[id] = struct{}{}
	}
	return cohortIds
}
