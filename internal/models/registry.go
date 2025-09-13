package models

import (
	"errors"
	"sync"
)

// Registry manages the registration and retrieval of model providers
type Registry struct {
	providers map[string]ModelProvider  // Uses ModelProvider from interface.go
	mu        sync.RWMutex
}

// NewRegistry creates a new model registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ModelProvider),
	}
}

// Register adds a new model provider to the registry
func (r *Registry) Register(name string, provider ModelProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return errors.New("provider name cannot be empty")
	}

	if provider == "" {
		return errors.New("provider cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return errors.New("provider already registered with this name")
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a model provider by name
func (r *Registry) Get(name string) (ModelProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return "", errors.New("provider not found")
	}

	return provider, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return errors.New("provider not found")
	}

	delete(r.providers, name)
	return nil
}

// Clear removes all providers from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = make(map[string]ModelProvider)
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}