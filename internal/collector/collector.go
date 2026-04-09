package collector

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"grpc-tracer/internal/storage"
)

// Collector collects trace spans from gRPC interceptors
type Collector struct {
	store       *storage.TraceStore
	serviceName string
}

// NewCollector creates a new trace collector
func NewCollector(serviceName string, store *storage.TraceStore) *Collector {
	return &Collector{
		store:       store,
		serviceName: serviceName,
	}
}

// CollectUnaryServerSpan collects span data from a unary server interceptor
func (c *Collector) CollectUnaryServerSpan(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	startTime := time.Now()

	// Extract trace context from metadata
	md, _ := metadata.FromIncomingContext(ctx)
	traceID := c.getMetadataValue(md, "x-trace-id")
	parentID := c.getMetadataValue(md, "x-span-id")
	spanID := c.generateSpanID()

	// Call the handler
	resp, err := handler(ctx, req)
	duration := time.Since(startTime)

	// Create span
	span := &storage.TraceSpan{
		TraceID:    traceID,
		SpanID:     spanID,
		ParentID:   parentID,
		Service:    c.serviceName,
		Method:     info.FullMethod,
		StartTime:  startTime,
		Duration:   duration,
		StatusCode: status.Code(err).String(),
		Metadata:   make(map[string]interface{}),
	}

	if err != nil {
		span.Error = err.Error()
	}

	// Store the span
	c.store.AddSpan(span)

	return resp, err
}

func (c *Collector) getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c *Collector) generateSpanID() string {
	return time.Now().Format("20060102150405.000000")
}

// GetStore returns the underlying trace store
func (c *Collector) GetStore() *storage.TraceStore {
	return c.store
}
