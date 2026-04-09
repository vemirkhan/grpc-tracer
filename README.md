# grpc-tracer

Lightweight CLI tool for tracing and visualizing gRPC call chains across microservices in real time.

## Features

- 🔍 Real-time gRPC call tracing across distributed services
- 📊 Visual call chain representation in your terminal
- ⚡ Minimal overhead and zero code changes required
- 🎯 Filter by service name, method, or trace ID
- 📝 Export traces to JSON for further analysis

## Installation

```bash
go install github.com/yourusername/grpc-tracer@latest
```

Or download pre-built binaries from the [releases page](https://github.com/yourusername/grpc-tracer/releases).

## Usage

Start tracing gRPC calls on a specific port:

```bash
grpc-tracer --port 50051
```

Trace with filters:

```bash
# Filter by service name
grpc-tracer --port 50051 --service user-service

# Filter by method
grpc-tracer --port 50051 --method GetUser

# Export to JSON
grpc-tracer --port 50051 --output traces.json
```

### Example Output

```
[2024-01-15 10:23:45] TRACE ID: abc123
├─ api-gateway.CreateOrder (2.3ms)
│  ├─ user-service.GetUser (1.1ms)
│  ├─ inventory-service.CheckStock (0.8ms)
│  └─ payment-service.ProcessPayment (15.2ms)
└─ TOTAL: 19.4ms
```

## Configuration

Set environment variables for advanced configuration:

- `GRPC_TRACER_BUFFER_SIZE`: Trace buffer size (default: 1000)
- `GRPC_TRACER_SAMPLE_RATE`: Sampling rate percentage (default: 100)

## License

MIT License - see [LICENSE](LICENSE) for details.