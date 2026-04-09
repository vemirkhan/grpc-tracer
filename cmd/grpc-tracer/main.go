// Package main is the entry point for the grpc-tracer CLI tool.
// It wires together all internal packages and exposes a command-line
// interface for tracing, visualizing, and exporting gRPC call chains.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/grpc-tracer/internal/circuitbreaker"
	"github.com/your-org/grpc-tracer/internal/collector"
	"github.com/your-org/grpc-tracer/internal/exporter"
	"github.com/your-org/grpc-tracer/internal/filter"
	"github.com/your-org/grpc-tracer/internal/ratelimiter"
	"github.com/your-org/grpc-tracer/internal/sampler"
	"github.com/your-org/grpc-tracer/internal/storage"
	"github.com/your-org/grpc-tracer/internal/visualizer"
)

// config holds all CLI flags parsed at startup.
type config struct {
	exportFormat   string
	exportOutput   string
	sampleRate     float64
	rateLimit      float64
	cbThreshold    int
	cbTimeout      time.Duration
	filterService  string
	filterErrors   bool
	filterMinMS    int64
	traceID        string
	listAll        bool
	version        bool
}

const appVersion = "0.1.0"

func parseFlags() config {
	var cfg config

	flag.StringVar(&cfg.exportFormat, "format", "text", "Export format: text or json")
	flag.StringVar(&cfg.exportOutput, "output", "", "Output file path (default: stdout)")
	flag.Float64Var(&cfg.sampleRate, "sample-rate", 1.0, "Sampling rate between 0.0 and 1.0")
	flag.Float64Var(&cfg.rateLimit, "rate-limit", 100.0, "Max spans accepted per second")
	flag.IntVar(&cfg.cbThreshold, "cb-threshold", 5, "Circuit breaker failure threshold")
	flag.DurationVar(&cfg.cbTimeout, "cb-timeout", 30*time.Second, "Circuit breaker open timeout")
	flag.StringVar(&cfg.filterService, "service", "", "Filter traces by service name")
	flag.BoolVar(&cfg.filterErrors, "errors-only", false, "Show only traces containing errors")
	flag.Int64Var(&cfg.filterMinMS, "min-duration-ms", 0, "Minimum trace duration in milliseconds")
	flag.StringVar(&cfg.traceID, "trace-id", "", "Display a specific trace by ID")
	flag.BoolVar(&cfg.listAll, "list-all", false, "List and visualize all stored traces")
	flag.BoolVar(&cfg.version, "version", false, "Print version and exit")

	flag.Parse()
	return cfg
}

func main() {
	cfg := parseFlags()

	if cfg.version {
		fmt.Printf("grpc-tracer version %s\n", appVersion)
		os.Exit(0)
	}

	// Initialise core components.
	store := storage.NewTraceStore()

	samp := sampler.New(sampler.Config{
		Rate: cfg.sampleRate,
	})

	rl := ratelimiter.New(ratelimiter.Config{
		Rate: cfg.rateLimit,
	})

	cb := circuitbreaker.New(circuitbreaker.Config{
		Threshold: cfg.cbThreshold,
		Timeout:   cfg.cbTimeout,
	})

	col := collector.NewCollector(store, samp, rl, cb)
	vis := visualizer.NewVisualizer(store)

	// Configure exporter.
	expOpts := []exporter.Option{
		exporter.WithFormat(cfg.exportFormat),
	}
	if cfg.exportOutput != "" {
		f, err := os.Create(cfg.exportOutput)
		if err != nil {
			log.Fatalf("grpc-tracer: cannot open output file: %v", err)
		}
		defer f.Close()
		expOpts = append(expOpts, exporter.WithWriter(f))
	}
	exp := exporter.NewExporter(store, expOpts...)

	// Suppress "declared and not used" errors for components used by the
	// gRPC server wiring (demonstrated here via a quick status print).
	log.Printf("grpc-tracer %s started (sampler=%.2f, rate-limit=%.0f/s, cb-threshold=%d)",
		appVersion, cfg.sampleRate, cfg.rateLimit, cfg.cbThreshold)
	_ = col // collector is consumed by interceptors in server wiring

	// Handle a specific trace lookup.
	if cfg.traceID != "" {
		handleTraceQuery(cfg, vis, exp)
		return
	}

	// List all traces and exit.
	if cfg.listAll {
		handleListAll(cfg, vis, exp)
		return
	}

	// Default: wait for OS signal (future: start embedded gRPC listener).
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	log.Println("grpc-tracer: waiting for traces — press Ctrl+C to stop")
	<-sig
	log.Println("grpc-tracer: shutting down")
}

func handleTraceQuery(cfg config, vis *visualizer.Visualizer, exp *exporter.Exporter) {
	if cfg.exportFormat == "json" {
		if err := exp.ExportTrace(cfg.traceID); err != nil {
			log.Fatalf("grpc-tracer: export failed: %v", err)
		}
		return
	}
	out, err := vis.FormatTrace(cfg.traceID)
	if err != nil {
		log.Fatalf("grpc-tracer: trace not found: %v", err)
	}
	fmt.Println(out)
}

func handleListAll(cfg config, vis *visualizer.Visualizer, exp *exporter.Exporter) {
	traces := vis.Store().GetAllTraces()

	// Apply filters when requested.
	if cfg.filterService != "" || cfg.filterErrors || cfg.filterMinMS > 0 {
		traces = filter.FilterTraces(traces, filter.Filter{
			ServiceName:   cfg.filterService,
			OnlyErrors:    cfg.filterErrors,
			MinDurationMS: cfg.filterMinMS,
		})
	}

	if len(traces) == 0 {
		log.Println("grpc-tracer: no traces match the given criteria")
		return
	}

	if cfg.exportFormat == "json" {
		if err := exp.ExportAll(); err != nil {
			log.Fatalf("grpc-tracer: export failed: %v", err)
		}
		return
	}

	out := vis.FormatAllTraces()
	fmt.Println(out)
}
