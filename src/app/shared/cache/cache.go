// Package cache provides memory cache for this application.
// For now cache implementation on zero stage - have only map[string interface] for caching info.
// If it need, we can upgrade it to https://github.com/patrickmn/go-cache or another production-ready cache
package cache

import (
	"sync"
)

const (
	// BestLenderRateCacheKey key for best lender lates from POST /api/admin/rate/list
	BestLenderRateCacheKey = "bestLenderRates"
)

var (
	cacheMap sync.Map // primitive key/value cache
)

// Put save data to cache
func Put(key string, data interface{}) {
	cacheMap.Store(key, data)
}

// Get return cached cached interface and bool variable that represents "find" state
func Get(key string) (interface{}, bool) {
	lenderList, ok := cacheMap.Load(key)
	if !ok {
		return nil, false
	}

	return lenderList, true
}
