package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

const permissionCacheKey = "permissions:matrix:v1"
const permissionCacheTTL = 10 * time.Minute

// PermissionRepository defines the interface for permission storage.
type PermissionRepository interface {
	GetAll(ctx context.Context) (map[string]map[string]bool, error)
	GetByRole(ctx context.Context, role string) (map[string]bool, error)
	SaveAll(ctx context.Context, perms map[string]map[string]bool) error
}

// PermissionService manages role-based permissions with caching.
type PermissionService struct {
	repo  PermissionRepository
	cache cache.Cache
	mu    sync.RWMutex
	// In-memory cache for fast lookups
	matrix map[string]map[string]bool
}

func NewPermissionService(repo PermissionRepository, c cache.Cache) *PermissionService {
	s := &PermissionService{
		repo:  repo,
		cache: c,
	}
	// Load permissions on startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.loadMatrix(ctx); err != nil {
		log.Printf("Warning: failed to load permissions on startup: %v", err)
	}
	return s
}

// HasPermission checks if a role has a specific permission.
func (s *PermissionService) HasPermission(role, permissionKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.matrix == nil {
		// If matrix not loaded, owner/admin always have access
		return role == "owner" || role == "admin"
	}

	perms, ok := s.matrix[role]
	if !ok {
		return false
	}
	return perms[permissionKey]
}

// GetAll returns the full permission matrix.
func (s *PermissionService) GetAll(ctx context.Context) (map[string]map[string]bool, error) {
	s.mu.RLock()
	if s.matrix != nil {
		defer s.mu.RUnlock()
		return s.matrix, nil
	}
	s.mu.RUnlock()

	if err := s.loadMatrix(ctx); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.matrix, nil
}

// GetByRole returns permissions for a specific role.
func (s *PermissionService) GetByRole(ctx context.Context, role string) (map[string]bool, error) {
	matrix, err := s.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	perms, ok := matrix[role]
	if !ok {
		return make(map[string]bool), nil
	}
	return perms, nil
}

// SaveAll saves the full permission matrix and invalidates cache.
func (s *PermissionService) SaveAll(ctx context.Context, perms map[string]map[string]bool) error {
	if err := s.repo.SaveAll(ctx, perms); err != nil {
		return err
	}

	// Update in-memory cache
	s.mu.Lock()
	s.matrix = perms
	s.mu.Unlock()

	// Update Redis cache
	data, err := json.Marshal(perms)
	if err == nil {
		_ = s.cache.Set(ctx, permissionCacheKey, data, permissionCacheTTL)
	}

	return nil
}

// loadMatrix loads the permission matrix from Redis cache or DB.
func (s *PermissionService) loadMatrix(ctx context.Context) error {
	// Try Redis first
	data, err := s.cache.Get(ctx, permissionCacheKey)
	if err == nil && data != nil {
		var matrix map[string]map[string]bool
		if json.Unmarshal(data, &matrix) == nil {
			s.mu.Lock()
			s.matrix = matrix
			s.mu.Unlock()
			return nil
		}
	}

	// Load from DB
	matrix, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.matrix = matrix
	s.mu.Unlock()

	// Cache to Redis
	data, err = json.Marshal(matrix)
	if err == nil {
		_ = s.cache.Set(ctx, permissionCacheKey, data, permissionCacheTTL)
	}

	return nil
}
