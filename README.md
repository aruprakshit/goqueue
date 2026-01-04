# GoQueue - Concurrent Job Processing System

A hands-on tutorial project for learning Go concurrency patterns by building a rate-limited concurrent task queue.

## Overview

GoQueue is a learning project that demonstrates Go's concurrency primitives through a practical job processing system. It covers goroutines, channels, worker pools, rate limiting, and more.

## Features

- **Producer** - Generates jobs at configurable intervals using Tickers
- **Worker Pool** - Multiple concurrent workers processing jobs
- **Rate Limiting** - Token bucket pattern to control throughput
- **Metrics Server** - Stateful goroutine tracking job statistics
- **Graceful Shutdown** - Clean exit with timeout handling

## Prerequisites

- Go 1.25 or higher

## Quick Start

```bash
# Clone and navigate to project
cd golang

# Run the application
go run .

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

## Project Structure

```
golang/
├── main.go              # Entry point, orchestrates all components
├── job/
│   └── job.go           # Job and Result types
├── producer/
│   └── producer.go      # Generates jobs using Ticker
├── consumer/
│   └── consumer.go      # Worker pool with rate limiting
├── metrics/
│   ├── metrics.go       # Atomic counters + Mutex approach
│   └── server.go        # Stateful goroutine approach
├── errors/
│   └── errors.go        # Custom error types
├── PATTERNS.md          # Detailed documentation of all patterns
└── README.md
```

## Configuration

Constants in `main.go`:

| Constant | Default | Description |
|----------|---------|-------------|
| `numJobs` | 5 | Number of jobs to process |
| `numWorkers` | 3 | Number of concurrent workers |
| `shutdownTimeout` | 5s | Max time before forced shutdown |
| `tickInterval` | 100ms | Time between job generation |
| `rateLimit` | 200ms | Rate limit per worker (5 jobs/sec) |

## Architecture

```
┌──────────────┐     jobs channel      ┌─────────────┐     results channel
│   Producer   │ ─────────────────────▶│ Worker Pool │ ────────────────────▶ Collector
│  (Ticker)    │    (buffered)         │ (3 workers) │   (non-blocking)
└──────────────┘                       │ +rate limit │
                                       └──────┬──────┘
                                              │
                                              ▼
                                    ┌──────────────────┐
                                    │  Metrics Server  │
                                    │ (stateful goroutine)
                                    └──────────────────┘
```

## Concepts Covered

This project is built in 7 phases, each introducing new concepts:

| Phase | Concepts |
|-------|----------|
| 1 | Goroutines, Channels, WaitGroups, Custom Errors |
| 2 | Worker Pools, Channel Directions, Range over Channels, Closing Channels |
| 3 | Select, Timeouts, Non-Blocking Operations, Timers |
| 4 | Tickers, Buffered Channels |
| 5 | Atomic Counters, Mutexes |
| 6 | Rate Limiting |
| 7 | Stateful Goroutines |

## Documentation

See [PATTERNS.md](PATTERNS.md) for detailed explanations of all 20 concurrency patterns used in this project, including:

- Code examples
- Visual diagrams
- Key points and best practices
- When to use each pattern

## Sample Output

```
=== Results ===
[Producer] Created Job{ID: 1, Name: task-1}
[Consumer 1] Processing Job{ID: 1, Name: task-1} (attempt 1/3)
[Consumer 1] Completed Job{ID: 1, Name: task-1} in 92ms
✓ Job 1: Processed payload: payload-for-job-1 (took 92ms)
...
[Main] All workers finished normally

=== Summary ===
Total: 5 | Completed: 5 | Failed: 0
Metrics: Processed: 5 | Failed: 0
```

## Testing

Each package includes tests:

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./job
go test ./producer
go test ./consumer
go test ./errors

# Run with race detector
go test -race ./...
```

## Related Resources

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go by Example](https://gobyexample.com/)

## License

MIT License - feel free to use this for learning purposes.
