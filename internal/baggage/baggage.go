// Package baggage provides utilities for attaching and retrieving
// arbitrary key-value pairs (baggage) to/from gRPC metadata, enabling
// propagation of contextual data across service boundaries.
package baggage

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

const prefix = "baggage-"

// Baggage holds a set of key-value pairs to propagate.
type Baggage map[string]string

// Inject writes all baggage entries into the outgoing gRPC metadata
// of the provided context. Keys are prefixed with "baggage-".
func Inject(ctx context.Context, b Baggage) context.Context {
	if len(b) == 0 {
		return ctx
	}
	pairs := make([]string, 0, len(b)*2)
	for k, v := range b {
		if k == "" || v == "" {
			continue
		}
		pairs = append(pairs, prefix+strings.ToLower(k), v)
	}
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

// Extract reads all baggage entries from the incoming gRPC metadata
// of the provided context. Only keys with the "baggage-" prefix are
// returned, with the prefix stripped.
func Extract(ctx context.Context) Baggage {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return Baggage{}
	}
	result := Baggage{}
	for k, vals := range md {
		if strings.HasPrefix(k, prefix) && len(vals) > 0 {
			bare := strings.TrimPrefix(k, prefix)
			result[bare] = vals[0]
		}
	}
	return result
}

// Get retrieves a single baggage value by key from incoming metadata.
// Returns an empty string if the key is not present.
func Get(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(prefix + strings.ToLower(key))
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
