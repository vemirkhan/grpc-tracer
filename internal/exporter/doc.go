// Package exporter provides functionality for exporting collected gRPC trace
// data to various output formats and destinations.
//
// Supported formats:
//   - JSON: structured JSON output suitable for ingestion by external tools
//   - Text: human-readable plain-text output for CLI inspection
//
// Usage:
//
//	store := storage.NewTraceStore()
//	exp := exporter.NewExporter(store, exporter.FormatJSON, os.Stdout)
//	if err := exp.ExportAll(); err != nil {
//	    log.Fatal(err)
//	}
package exporter
