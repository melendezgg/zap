package main

import "sync"

type RouteInfo struct {
	File    string
	Type    string
	IsTS    bool
	CSSFile string
	Title   string
}

type RouteStore struct {
	mu     sync.RWMutex
	routes map[string]*RouteInfo
}

func NewRouteStore() *RouteStore {
	return &RouteStore{
		routes: make(map[string]*RouteInfo),
	}
}

func (s *RouteStore) Get(path string) (*RouteInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.routes[path]
	return info, ok
}

func (s *RouteStore) SetAll(routes map[string]*RouteInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routes = routes
}

func (s *RouteStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.routes)
}

var routeStore = NewRouteStore()

type BundleCache struct {
	mu      sync.RWMutex
	bundles map[string]string
}

func NewBundleCache() *BundleCache {
	return &BundleCache{
		bundles: make(map[string]string),
	}
}

func (c *BundleCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bundle, ok := c.bundles[key]
	return bundle, ok
}

func (c *BundleCache) Set(key, bundle string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bundles[key] = bundle
}

func (c *BundleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bundles = make(map[string]string)
}

var bundleCache = NewBundleCache()
