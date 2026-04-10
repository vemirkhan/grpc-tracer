// Package tracegateway provides a lightweight HTTP gateway for exposing
// trace data collected by grpc-tracer over a REST-like API.
//
// # Endpoints
//
//	GET /traces        – list all traces (returns JSON with count + trace list)
//	GET /traces/{id}   – fetch spans for a single trace by ID
//	GET /health        – liveness check, always returns {"status":"ok"}
//
// # Middleware
//
// LoggingMiddleware and CORSMiddleware are provided and can be composed
// using the Chain helper:
//
//	gw := tracegateway.New(store, vis)
//	handler := tracegateway.Chain(gw,
//	    tracegateway.LoggingMiddleware,
//	    tracegateway.CORSMiddleware,
//	)
//	http.ListenAndServe(":8080", handler)
package tracegateway
