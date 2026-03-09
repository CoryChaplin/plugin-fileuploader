package server

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

// adsTxtCache fetches and caches an ads.txt file from a remote URL,
// refreshing the content daily in the background.
type adsTxtCache struct {
	mu      sync.RWMutex
	content []byte
	url     string
}

// newAdsTxtCache starts a cache for the given URL. It fetches the content
// immediately on startup, then spawns a goroutine to refresh every 24 hours.
// The goroutine stops when ctx is cancelled.
func newAdsTxtCache(ctx context.Context, url string) *adsTxtCache {
	c := &adsTxtCache{url: url}
	c.refresh()
	go c.updateLoop(ctx)
	return c
}

func (c *adsTxtCache) refresh() {
	resp, err := http.Get(c.url) //nolint:noctx
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.content = body
	c.mu.Unlock()
}

func (c *adsTxtCache) updateLoop(ctx context.Context) {
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.refresh()
		case <-ctx.Done():
			return
		}
	}
}

// Get returns the cached ads.txt content, or nil if it has not been loaded yet.
func (c *adsTxtCache) Get() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.content
}
