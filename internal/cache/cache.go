// Package cache provides an in-memory caching mechanism using the go-cache library.
package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Eager(
		cache.New(15*time.Minute, 20*time.Minute),
	),
)
