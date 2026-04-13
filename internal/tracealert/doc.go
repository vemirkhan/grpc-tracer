// Package tracealert implements threshold-based alerting for gRPC trace spans.
//
// An Alerter is configured with duration thresholds and an error-sensitivity
// flag. Each call to Evaluate inspects a single span and records an Alert
// whenever a threshold is breached:
//
//   - WarnDuration  — span took longer than the warn threshold.
//   - ErrorDuration — span took longer than the error threshold.
//   - AlertOnError  — span carries a non-empty error string.
//
// Alerts are stored in memory and can be retrieved with All or cleared with
// Clear. The package is safe for concurrent use.
//
// Example:
//
//	alerter := tracealert.New(tracealert.Config{
//		WarnDuration:  500 * time.Millisecond,
//		ErrorDuration: 2 * time.Second,
//		AlertOnError:  true,
//	})
//	alerter.Evaluate(span)
//	for _, a := range alerter.All() {
//		fmt.Println(a.Level, a.Message)
//	}
package tracealert
