package rbac

import (
	"context"
	"sync"
	"time"

	"github.com/quickbite/analytics-service/lib/coreclient"
)

type PermissionCache struct {
	client *coreclient.Client
	ttl    time.Duration
	mu     sync.RWMutex
	cache  map[string]cacheEntry
}

type cacheEntry struct {
	permissions []string
	cachedAt    time.Time
}

func NewPermissionCache(client *coreclient.Client, ttl time.Duration) *PermissionCache {
	return &PermissionCache{
		client: client,
		ttl:    ttl,
		cache:  make(map[string]cacheEntry),
	}
}

func (pc *PermissionCache) GetPermissions(ctx context.Context, roleName string) ([]string, error) {
	pc.mu.RLock()
	entry, ok := pc.cache[roleName]
	pc.mu.RUnlock()

	if ok && time.Since(entry.cachedAt) < pc.ttl {
		return entry.permissions, nil
	}

	perms, err := pc.client.GetRolePermissions(ctx, roleName)
	if err != nil {
		return nil, err
	}

	pc.mu.Lock()
	pc.cache[roleName] = cacheEntry{permissions: perms, cachedAt: time.Now()}
	pc.mu.Unlock()

	return perms, nil
}

func (pc *PermissionCache) HasPermission(permissions []string, perm string) bool {
	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func (pc *PermissionCache) Invalidate(roleName string) {
	pc.mu.Lock()
	delete(pc.cache, roleName)
	pc.mu.Unlock()
}
