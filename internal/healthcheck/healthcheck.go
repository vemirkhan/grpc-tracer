// Package healthcheck provides a lightweight health-check probe for gRPC
// services participating in the tracer ecosystem. It records liveness and
// readiness signals and exposes them for CLI or HTTP consumption.
package healthcheck

import (
	"sync"
	"time"
)

// Status represents the health state of a service.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ServiceHealth holds the latest health snapshot for a single service.
type ServiceHealth struct {
	ServiceName string    `json:"service_name"`
	Status      Status    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Message     string    `json:"message,omitempty"`
}

// Checker stores and updates health records for tracked services.
type Checker struct {
	mu       sync.RWMutex
	services map[string]*ServiceHealth
}

// New returns an initialised Checker.
func New() *Checker {
	return &Checker{
		services: make(map[string]*ServiceHealth),
	}
}

// Record upserts the health status for a named service.
func (c *Checker) Record(name string, s Status, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = &ServiceHealth{
		ServiceName: name,
		Status:      s,
		LastChecked: time.Now().UTC(),
		Message:     msg,
	}
}

// Get returns the health record for a service, or false if unknown.
func (c *Checker) Get(name string) (ServiceHealth, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	h, ok := c.services[name]
	if !ok {
		return ServiceHealth{}, false
	}
	return *h, true
}

// All returns a snapshot of every tracked service.
func (c *Checker) All() []ServiceHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]ServiceHealth, 0, len(c.services))
	for _, h := range c.services {
		out = append(out, *h)
	}
	return out
}

// IsHealthy returns true only when every tracked service is StatusHealthy.
func (c *Checker) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.services {
		if h.Status != StatusHealthy {
			return false
		}
	}
	return true
}
