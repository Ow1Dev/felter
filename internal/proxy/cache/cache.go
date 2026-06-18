// Package cache provides user identity caching for the proxy.
package cache

import "context"

// UserCache stores and retrieves mappings from (provider, providerID) to user ID.
type UserCache interface {
	// Get returns the cached user ID for the given provider and provider ID.
	// found is false when the key is not present or has expired.
	Get(ctx context.Context, provider, providerID string) (userID int64, found bool, err error)
	// Set stores a mapping from (provider, providerID) to user ID.
	Set(ctx context.Context, provider, providerID string, userID int64) error
	// Close releases any resources held by the cache.
	Close()
}
