// Package logger provides structured, level-based logging for gRPC spans
// captured by grpc-tracer.
//
// # Overview
//
// A Logger writes one log line per Span to any io.Writer.  The severity level
// is derived automatically from the span's state:
//
//   - ERROR  — the span recorded a non-empty error string
//   - WARN   — the span duration exceeded 500 ms
//   - INFO   — all other spans
//
// # Usage
//
//	l := logger.New(os.Stdout, logger.LevelInfo)
//
//	// log a single span
//	l.LogSpan(span)
//
//	// log every span in a trace
//	l.LogTrace(spans)
//
//	// attach as a gRPC server interceptor
//	grpc.NewServer(
//		grpc.UnaryInterceptor(logger.UnaryServerInterceptor(l, "my-service")),
//	)
package logger
