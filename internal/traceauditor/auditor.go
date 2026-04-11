// Package traceauditor records access and mutation events on traces for
// audit and compliance purposes.
package traceauditor

import (
	"fmt"
	"sync"
	"time"
)

// EventKind describes the type of audit event.
type EventKind string

const (
	EventRead   EventKind = "read"
	EventWrite  EventKind = "write"
	EventDelete EventKind = "delete"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time
	Kind      EventKind
	TraceID   string
	SpanID    string
	Actor     string
	Detail    string
}

func (e Event) String() string {
	return fmt.Sprintf("[%s] %s actor=%s trace=%s span=%s detail=%s",
		e.Timestamp.Format(time.RFC3339), e.Kind, e.Actor, e.TraceID, e.SpanID, e.Detail)
}

// Auditor records audit events in memory.
type Auditor struct {
	mu     sync.RWMutex
	events []Event
}

// New creates a new Auditor.
func New() *Auditor {
	return &Auditor{}
}

// Record appends a new audit event.
func (a *Auditor) Record(kind EventKind, traceID, spanID, actor, detail string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, Event{
		Timestamp: time.Now(),
		Kind:      kind,
		TraceID:   traceID,
		SpanID:    spanID,
		Actor:     actor,
		Detail:    detail,
	})
}

// All returns a snapshot of all recorded events.
func (a *Auditor) All() []Event {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Event, len(a.events))
	copy(out, a.events)
	return out
}

// FilterByTrace returns events associated with the given traceID.
func (a *Auditor) FilterByTrace(traceID string) []Event {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var out []Event
	for _, e := range a.events {
		if e.TraceID == traceID {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the total number of recorded events.
func (a *Auditor) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.events)
}
