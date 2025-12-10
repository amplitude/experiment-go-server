package local

import (
	"github.com/amplitude/experiment-go-server/internal/cache"
	"sync"
)

type exposureFilter struct {
	mu        sync.Mutex
	cache     *cache.Cache
	ttlMillis int64
}

func newExposureFilter(size int) *exposureFilter {
	filter := &exposureFilter{
		cache:     cache.NewCache(size, dayMillis),
		ttlMillis: dayMillis,
	}
	return filter
}

func (f *exposureFilter) shouldTrack(exposure *exposure) bool {
	if len(exposure.results) == 0 {
		// Don't track empty exposures.
		return false
	}
	canonicalExposure := exposure.Canonicalize()
	f.mu.Lock()
	defer f.mu.Unlock()
	_, found := f.cache.Get(canonicalExposure)
	if !found {
		f.cache.Set(canonicalExposure, nil)
		return true
	}
	return false
}
