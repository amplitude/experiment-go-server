package local

import (
	"github.com/amplitude/experiment-go-server/internal/cache"
	"sync"
)

type AssignmentFilter struct {
	mu    sync.Mutex
	cache *cache.Cache
}

func NewAssignmentFilter(size int) *AssignmentFilter {

	filter := &AssignmentFilter{
		cache: cache.NewCache(size, DayMillis),
	}
	return filter
}

func (f *AssignmentFilter) shouldTrack(assignment *Assignment) bool {
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
