// Package provider contains data retriving and cache logic.
// For example of GetBestLenderRatesList:
// Controller request info not from DB or Cache, but from provider. Provider have access to database,
// cache or third party services. So, provider returns requested info from DB or cache, or another place,
// and frees the controller from unnecessary logic.
// Ideally, all data retriving logic should be in provider.
package provider

import (
	"app/constants"
	"app/shared/cache"

	"log"

	"github.com/jasonlvhit/gocron"
)

const (
	cacheInUse = true // enable/disable caching in app
)

func init() {
	if !cacheInUse {
		log.Println("WARNING: cache disabled")
	}
}

// InitAppCache add some init values to cache
func InitAppCache() {
	_, err := getBestRatesListFromDB()
	if err != nil {
		panic(err)
	}

	startUpdateCacheSheduler()
}

// UpdateCache do cache updating by command
func UpdateCache(command string) error {
	switch command {
	case "update_best_rates":
		rateList, err := getBestRatesListFromDB()
		if err != nil {
			return err
		}

		cache.Put(cache.BestLenderRateCacheKey, rateList)

	default:
		return constants.NewErrorBadParams("command void or not implemented command")
	}

	return nil
}

// StartUpdateCacheSheduler start gocron with update cache task
func startUpdateCacheSheduler() {
	log.Println("start cache update sheduler")

	gocron.Every(1).Day().Do(func() {
		InitAppCache()
	})
}
