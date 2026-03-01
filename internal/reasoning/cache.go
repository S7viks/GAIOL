package reasoning

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CachedResponse represents a cached model output with expiration
type CachedResponse struct {
	Output    ModelOutput
	CachedAt  time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (cr *CachedResponse) IsExpired() bool {
	return time.Now().After(cr.ExpiresAt)
}

// ResponseCache provides thread-safe in-memory caching for model responses
type ResponseCache struct {
	cache map[string]CachedResponse
	mu    sync.RWMutex
	ttl   time.Duration // Time to live for cache entries
}

// NewResponseCache creates a new response cache with specified TTL
func NewResponseCache(ttl time.Duration) *ResponseCache {
	if ttl == 0 {
		ttl = 1 * time.Hour // Default 1 hour
	}

	cache := &ResponseCache{
		cache: make(map[string]CachedResponse),
		ttl:   ttl,
	}

	// Start cleanup goroutine to remove expired entries
	go cache.cleanupExpired()

	return cache
}

// generateCacheKey creates a deterministic hash key from step parameters
func (rc *ResponseCache) generateCacheKey(objective, taskType, context string) string {
	// Combine all inputs that affect the response
	combined := fmt.Sprintf("%s|%s|%s", objective, taskType, context)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached response if it exists and hasn't expired
func (rc *ResponseCache) Get(objective, taskType, context string) (ModelOutput, bool) {
	key := rc.generateCacheKey(objective, taskType, context)

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	cached, exists := rc.cache[key]
	if !exists || cached.IsExpired() {
		return ModelOutput{}, false
	}

	return cached.Output, true
}

// Set stores a model output in the cache
func (rc *ResponseCache) Set(objective, taskType, context string, output ModelOutput) {
	key := rc.generateCacheKey(objective, taskType, context)

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache[key] = CachedResponse{
		Output:    output,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(rc.ttl),
	}
}

// cleanupExpired periodically removes expired cache entries
func (rc *ResponseCache) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rc.mu.Lock()
		for key, cached := range rc.cache {
			if cached.IsExpired() {
				delete(rc.cache, key)
			}
		}
		rc.mu.Unlock()
	}
}

// Stats returns cache statistics
func (rc *ResponseCache) Stats() map[string]int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	totalEntries := len(rc.cache)
	expiredCount := 0

	for _, cached := range rc.cache {
		if cached.IsExpired() {
			expiredCount++
		}
	}

	return map[string]int{
		"total_entries":   totalEntries,
		"active_entries":  totalEntries - expiredCount,
		"expired_entries": expiredCount,
	}
}

// Clear removes all cache entries
func (rc *ResponseCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[string]CachedResponse)
}
