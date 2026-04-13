// Package tracealert provides threshold-based alerting over trace spans.
// Alerts are fired when span metrics exceed configured thresholds.
package tracealert

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Alert describes a fired alert.
type Alert struct {
	TraceID   string
	SpanID    string
	Service   string
	Level     Level
	Message   string
	FiredAt   time.Time
}

// Config holds alerting thresholds.
type Config struct {
	// WarnDuration fires a warn-level alert when a span exceeds this duration.
	WarnDuration time.Duration
	// ErrorDuration fires an error-level alert when a span exceeds this duration.
	ErrorDuration time.Duration
	// AlertOnError fires an error-level alert whenever a span has an error.
	AlertOnError bool
}

func defaultConfig() Config {
	return Config{
		WarnDuration:  500 * time.Millisecond,
		ErrorDuration: 2 * time.Second,
		AlertOnError:  true,
	}
}

// Alerter evaluates spans and collects fired alerts.
type Alerter struct {
	cfg    Config
	mu     sync.Mutex
	alerts []Alert
}

// New creates an Alerter. A zero-value Config applies sensible defaults.
func New(cfg Config) *Alerter {
	def := defaultConfig()
	if cfg.WarnDuration == 0 {
		cfg.WarnDuration = def.WarnDuration
	}
	if cfg.ErrorDuration == 0 {
		cfg.ErrorDuration = def.ErrorDuration
	}
	return &Alerter{cfg: cfg}
}

// Evaluate checks a span against the configured thresholds and records any alerts.
func (a *Alerter) Evaluate(span storage.Span) {
	var fired []Alert

	if a.cfg.AlertOnError && span.Error != "" {
		fired = append(fired, Alert{
			TraceID: span.TraceID,
			SpanID:  span.SpanID,
			Service: span.Service,
			Level:   LevelError,
			Message: fmt.Sprintf("span error: %s", span.Error),
			FiredAt: time.Now(),
		})
	}

	if span.Duration >= a.cfg.ErrorDuration {
		fired = append(fired, Alert{
			TraceID: span.TraceID,
			SpanID:  span.SpanID,
			Service: span.Service,
			Level:   LevelError,
			Message: fmt.Sprintf("span duration %s exceeds error threshold %s", span.Duration, a.cfg.ErrorDuration),
			FiredAt: time.Now(),
		})
	} else if span.Duration >= a.cfg.WarnDuration {
		fired = append(fired, Alert{
			TraceID: span.TraceID,
			SpanID:  span.SpanID,
			Service: span.Service,
			Level:   LevelWarn,
			Message: fmt.Sprintf("span duration %s exceeds warn threshold %s", span.Duration, a.cfg.WarnDuration),
			FiredAt: time.Now(),
		})
	}

	if len(fired) > 0 {
		a.mu.Lock()
		a.alerts = append(a.alerts, fired...)
		a.mu.Unlock()
	}
}

// All returns a snapshot of all fired alerts.
func (a *Alerter) All() []Alert {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]Alert, len(a.alerts))
	copy(out, a.alerts)
	return out
}

// Clear removes all stored alerts.
func (a *Alerter) Clear() {
	a.mu.Lock()
	a.alerts = nil
	a.mu.Unlock()
}
