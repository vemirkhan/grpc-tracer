// Package tracetagger provides automatic tag enrichment for spans based on
// configurable rules that match span attributes and apply key-value tags.
package tracetagger

import (
	"strings"
	"sync"

	"github.com/user/grpc-tracer/internal/storage"
)

// Rule defines a matching condition and the tags to apply when matched.
type Rule struct {
	// ServicePrefix matches spans whose service name starts with this value.
	// Empty string matches all services.
	ServicePrefix string
	// MethodContains matches spans whose method contains this substring.
	// Empty string matches all methods.
	MethodContains string
	// Tags are the key-value pairs applied when the rule matches.
	Tags map[string]string
}

// Tagger applies tag rules to spans in a trace store.
type Tagger struct {
	mu    sync.RWMutex
	rules []Rule
	store *storage.TraceStore
}

// New creates a Tagger backed by the given store.
func New(store *storage.TraceStore) (*Tagger, error) {
	if store == nil {
		return nil, errNilStore
	}
	return &Tagger{store: store}, nil
}

// AddRule appends a tagging rule to the tagger.
func (t *Tagger) AddRule(r Rule) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rules = append(t.rules, r)
}

// TagSpan applies all matching rules to the given span and persists the updated
// span back to the store. It returns the number of tags applied.
func (t *Tagger) TagSpan(span storage.Span) (int, error) {
	t.mu.RLock()
	rules := make([]Rule, len(t.rules))
	copy(rules, t.rules)
	t.mu.RUnlock()

	applied := 0
	for _, r := range rules {
		if !t.matches(r, span) {
			continue
		}
		if span.Tags == nil {
			span.Tags = make(map[string]string)
		}
		for k, v := range r.Tags {
			span.Tags[k] = v
			applied++
		}
	}

	if applied > 0 {
		if err := t.store.AddSpan(span); err != nil {
			return 0, err
		}
	}
	return applied, nil
}

func (t *Tagger) matches(r Rule, span storage.Span) bool {
	if r.ServicePrefix != "" && !strings.HasPrefix(span.Service, r.ServicePrefix) {
		return false
	}
	if r.MethodContains != "" && !strings.Contains(span.Method, r.MethodContains) {
		return false
	}
	return true
}

var errNilStore = errorString("tracetagger: store must not be nil")

type errorString string

func (e errorString) Error() string { return string(e) }
