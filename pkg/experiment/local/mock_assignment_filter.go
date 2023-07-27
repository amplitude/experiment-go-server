package local

import (
	"github.com/amplitude/experiment-go-server/internal/cache"
	"time"
)

type MockAssignmentFilter struct {
	cache *cache.Cache
}

func NewMockAssignmentFilter(size int, ttlMillis int) *AssignmentFilter {
	filter := &AssignmentFilter{
		cache: cache.NewCache(size, int64(ttlMillis)),
	}
	return filter
}

func (f *MockAssignmentFilter) shouldTrack(assignment *Assignment) bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	canonicalAssignment := assignment.Canonicalize()
	track, found := f.cache.Get(canonicalAssignment)
	if !found {
		f.cache.Set(canonicalAssignment, now)
		return true
	}
	return track == 0
}
