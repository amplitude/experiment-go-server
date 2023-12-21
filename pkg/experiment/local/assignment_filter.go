package local

import (
	"github.com/amplitude/experiment-go-server/internal/cache"
	"sync"
)

type assignmentFilter struct {
	mu    sync.Mutex
	cache *cache.Cache
}

func newAssignmentFilter(size int) *assignmentFilter {

	filter := &assignmentFilter{
		cache: cache.NewCache(size, dayMillis),
	}
	return filter
}

func (f *assignmentFilter) shouldTrack(assignment *assignment) bool {
	if len(assignment.results) == 0 {
		return false
	}
	canonicalAssignment := assignment.Canonicalize()
	f.mu.Lock()
	track, found := f.cache.Get(canonicalAssignment)
	if !found {
		f.cache.Set(canonicalAssignment, nil)
		f.mu.Unlock()
		return true
	}
	f.mu.Unlock()
	return track == 0
}
