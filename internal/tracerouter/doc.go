// Package tracerouter provides span-level routing for the grpc-tracer pipeline.
//
// A Router evaluates a set of user-defined Rule functions against each
// incoming Span and forwards the span to every matching destination
// TraceStore. When no rule matches, an optional fallback store receives
// the span instead.
//
// Typical usage:
//
//	auth := storage.NewTraceStore()
//	defaultStore := storage.NewTraceStore()
//
//	r := tracerouter.New()
//	_ = r.AddRoute(func(s storage.Span) bool {
//	    return s.Service == "auth-service"
//	}, auth)
//	r.SetFallback(defaultStore)
//
//	// Wire into gRPC server
//	grpc.NewServer(
//	    grpc.UnaryInterceptor(tracerouter.UnaryServerInterceptor(r)),
//	)
//
// The interceptor is safe for concurrent use.
package tracerouter
