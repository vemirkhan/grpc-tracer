// Package traceclassifier assigns human-readable labels to spans based on
// configurable rules evaluated against service name, method, duration, and tags.
package traceclassifier

import (
	"strings"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Label is a classification tag applied to a span.
type Label string

const (
	LabelFast    Label = "fast"
	LabelSlow    Label = "slow"
	LabelError   Label = "error"
	LabelInternal Label = "internal"
	LabelExternal Label = "external"
)

// Rule is a predicate that returns a Label when matched.
type Rule struct {
	Name    string
	Label   Label
	Match   func(span storage.Span) bool
}

// Classifier holds an ordered list of rules and applies them to spans.
type Classifier struct {
	rules []Rule
}

// New returns a Classifier pre-loaded with the supplied rules.
// Rules are evaluated in order; all matching rules contribute their label.
func New(rules ...Rule) *Classifier {
	return &Classifier{rules: rules}
}

// Classify returns all labels that match the given span.
func (c *Classifier) Classify(span storage.Span) []Label {
	var labels []Label
	seen := map[Label]bool{}
	for _, r := range c.rules {
		if r.Match(span) {
			if !seen[r.Label] {
				labels = append(labels, r.Label)
				seen[r.Label] = true
			}
		}
	}
	return labels
}

// DefaultRules returns a sensible set of built-in classification rules.
func DefaultRules(slowThreshold time.Duration) []Rule {
	if slowThreshold <= 0 {
		slowThreshold = 500 * time.Millisecond
	}
	return []Rule{
		{
			Name:  "error",
			Label: LabelError,
			Match: func(s storage.Span) bool { return s.Error != "" },
		},
		{
			Name:  "slow",
			Label: LabelSlow,
			Match: func(s storage.Span) bool { return s.Duration >= slowThreshold },
		},
		{
			Name:  "fast",
			Label: LabelFast,
			Match: func(s storage.Span) bool { return s.Duration < slowThreshold && s.Error == "" },
		},
		{
			Name:  "internal",
			Label: LabelInternal,
			Match: func(s storage.Span) bool {
				return strings.HasSuffix(s.ServiceName, "-internal") ||
					s.Tags["span.kind"] == "internal"
			},
		},
		{
			Name:  "external",
			Label: LabelExternal,
			Match: func(s storage.Span) bool {
				return s.Tags["span.kind"] == "client" || s.Tags["span.kind"] == "producer"
			},
		},
	}
}
