package healthcheck

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC server interceptor that records the
// health of the service after every RPC call. A successful call marks the
// service as StatusHealthy; a non-nil error marks it StatusDegraded or
// StatusUnhealthy depending on the gRPC status code.
func UnaryServerInterceptor(checker *Checker, serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		_ = start

		if err != nil {
			st, _ := status.FromError(err)
			msg := st.Message()
			if isServerFault(st.Code()) {
				checker.Record(serviceName, StatusUnhealthy, msg)
			} else {
				checker.Record(serviceName, StatusDegraded, msg)
			}
			return resp, err
		}

		checker.Record(serviceName, StatusHealthy, "")
		return resp, nil
	}
}

// isServerFault returns true for gRPC codes that indicate an internal server
// problem rather than a client-side issue.
func isServerFault(code interface{ String() string }) bool {
	switch code.String() {
	case "Internal", "Unavailable", "DataLoss", "Unknown":
		return true
	}
	return false
}
