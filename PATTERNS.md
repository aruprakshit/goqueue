# Go Concurrency Patterns - Learning Reference

This document captures all the patterns and techniques learned while building the GoQueue concurrent job processing system.

---

## Table of Contents

1. [Goroutines](#1-goroutines)
2. [Channels](#2-channels)
3. [Channel Directions](#3-channel-directions)
4. [WaitGroups](#4-waitgroups)
5. [Worker Pools](#5-worker-pools)
6. [Loop Variable Capture Gotcha](#6-loop-variable-capture-gotcha)
7. [Select Statement](#7-select-statement)
8. [Signal Channels with struct{}](#8-signal-channels-with-struct)
9. [Close to Broadcast Pattern](#9-close-to-broadcast-pattern)
10. [Timeouts with time.After](#10-timeouts-with-timeafter)
11. [Timers with time.NewTimer](#11-timers-with-timenewtimer)
12. [Non-Blocking Channel Operations](#12-non-blocking-channel-operations)
13. [Retry Logic Pattern](#13-retry-logic-pattern)
14. [Tickers with time.NewTicker](#14-tickers-with-timenewticker)
15. [Buffered Channels for Burst Handling](#15-buffered-channels-for-burst-handling)
16. [Multiple Defer Statements](#16-multiple-defer-statements)
17. [Atomic Counters](#17-atomic-counters)
18. [Mutexes](#18-mutexes)
19. [Rate Limiting](#19-rate-limiting)
20. [Stateful Goroutines](#20-stateful-goroutines)

---

## 1. Goroutines

Goroutines are lightweight threads managed by the Go runtime.

```go
// Start a goroutine
go func() {
    // This runs concurrently
    fmt.Println("Hello from goroutine")
}()
```

**Key points:**
- Prefix any function call with `go` to run it concurrently
- Goroutines are cheap (~2KB stack), you can spawn thousands
- The main function won't wait for goroutines to finish (use WaitGroups)

---

## 2. Channels

Channels are typed conduits for communication between goroutines.

```go
// Unbuffered channel - sender blocks until receiver is ready
ch := make(chan int)

// Buffered channel - sender blocks only when buffer is full
ch := make(chan int, 10)

// Send
ch <- 42

// Receive
value := <-ch

// Close (signals no more values)
close(ch)
```

**Key points:**
- Unbuffered channels synchronize sender and receiver
- Buffered channels allow async sends up to buffer size
- Only the sender should close a channel
- Receiving from a closed channel returns zero value immediately

---

## 3. Channel Directions

Restrict channel usage in function signatures for safety.

```go
// Send-only channel
func producer(jobs chan<- Job) {
    jobs <- Job{} // OK
    // <-jobs      // Compile error!
}

// Receive-only channel
func consumer(jobs <-chan Job) {
    j := <-jobs   // OK
    // jobs <- Job{} // Compile error!
}
```

**Key points:**
- `chan<-` = send-only (arrow points into channel)
- `<-chan` = receive-only (arrow points out of channel)
- Prevents accidental misuse at compile time

---

## 4. WaitGroups

Coordinate multiple goroutines - wait for all to finish.

```go
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Do work
    }()
}

wg.Wait() // Blocks until all Done() calls
```

**Key points:**
- `Add(n)` - increment counter before starting goroutine
- `Done()` - decrement counter (usually with defer)
- `Wait()` - block until counter reaches zero

---

## 5. Worker Pools

Multiple goroutines competing for work from a shared channel.

```go
jobs := make(chan Job)
numWorkers := 3

for i := 1; i <= numWorkers; i++ {
    go func(workerID int) {
        for job := range jobs {
            // Process job
            // Each job goes to exactly ONE worker
        }
    }(i)
}

// Send jobs - automatically distributed
jobs <- Job{ID: 1}
jobs <- Job{ID: 2}
close(jobs) // Signal no more jobs
```

**Key points:**
- Go automatically distributes work - one job per worker
- Workers exit when channel is closed and drained
- Use `for range` to receive until channel closes

---

## 6. Loop Variable Capture Gotcha

Classic bug when using loop variables in goroutines.

```go
// WRONG - all goroutines see final value of i
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i) // Prints 3, 3, 3
    }()
}

// CORRECT - pass as parameter
for i := 0; i < 3; i++ {
    go func(id int) {
        fmt.Println(id) // Prints 0, 1, 2 (order may vary)
    }(i) // Pass i as argument
}
```

**Key points:**
- Loop variable is shared across iterations
- By the time goroutine runs, loop may have finished
- Fix: pass variable as function parameter to capture current value

---

## 7. Select Statement

Listen on multiple channels simultaneously.

```go
select {
case job := <-jobs:
    // Received a job
case <-done:
    // Shutdown signal received
    return
case <-time.After(5 * time.Second):
    // Timeout
}
```

**Key points:**
- Blocks until ONE case is ready
- If multiple ready, picks randomly (fair)
- Use for multiplexing channel operations

---

## 8. Signal Channels with struct{}

Use empty struct for signal-only channels.

```go
done := make(chan struct{})

// Signal by closing
close(done)

// Wait for signal
<-done
```

**Why struct{} instead of bool?**
1. **Zero memory allocation** - `struct{}` is 0 bytes, `bool` is 1 byte
2. **Self-documenting** - signals "only the event matters, not the value"
3. **Idiomatic Go** - standard pattern for broadcast signals via close()

---

## 9. Close to Broadcast Pattern

Closing a channel unblocks ALL receivers simultaneously.

```go
done := make(chan struct{})

// Multiple listeners
for i := 0; i < 3; i++ {
    go func(id int) {
        <-done // All unblock when done is closed
        fmt.Printf("Worker %d stopping\n", id)
    }(i)
}

// Broadcast to all
close(done) // All 3 workers wake up
```

**Key points:**
- Sending only wakes ONE receiver
- Closing wakes ALL receivers
- Use for shutdown signals, cancellation

---

## 10. Timeouts with time.After

Add timeout to any channel operation.

```go
select {
case result := <-resultCh:
    // Got result in time
case <-time.After(150 * time.Millisecond):
    // Timeout - took too long
    return errors.New("operation timed out")
}
```

**Key points:**
- `time.After` returns a channel that receives after duration
- Combine with select for timeout behavior
- Creates a new timer each call (use `time.NewTimer` for reuse)

---

## 11. Timers with time.NewTimer

One-shot timer with ability to stop/reset.

```go
timer := time.NewTimer(2 * time.Second)

select {
case <-allDone:
    timer.Stop() // Cancel timer - cleanup
    fmt.Println("Finished before deadline")
case <-timer.C:
    fmt.Println("Deadline exceeded")
    close(done) // Signal shutdown
}
```

**Key points:**
- `timer.C` - channel that receives when timer fires
- `timer.Stop()` - cancel the timer (returns true if stopped before firing)
- `timer.Reset(d)` - reuse timer with new duration
- Prefer over `time.After` when you need to cancel

---

## 12. Non-Blocking Channel Operations

Check channel state without waiting using `default`.

```go
// Non-blocking send
select {
case results <- result:
    // Sent successfully
default:
    // Channel full - would block, so skip
    fmt.Println("WARNING: channel full, dropping message")
}

// Non-blocking receive
select {
case msg := <-messages:
    // Got a message
default:
    // No message available right now
}
```

**Key points:**
- `default` runs immediately if no channel is ready
- Prevents goroutine from blocking
- Useful for "try" semantics, backpressure handling

---

## 13. Retry Logic Pattern

Retry failed operations with attempt tracking.

```go
maxRetries := 3
var lastErr error

for attempt := 1; attempt <= maxRetries; attempt++ {
    err := doWork()
    if err == nil {
        return nil // Success
    }

    lastErr = err
    fmt.Printf("Attempt %d failed, retrying...\n", attempt)

    // Optional: exponential backoff
    // time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
}

return fmt.Errorf("all %d retries failed: %w", maxRetries, lastErr)
```

**Key points:**
- Track attempt count and last error
- Return immediately on success
- Consider adding backoff between retries
- Wrap original error for debugging

---

## 14. Tickers with time.NewTicker

Periodic events at fixed intervals (unlike Timer which fires once).

```go
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop() // Always clean up!

count := 0
for range ticker.C {
    count++
    fmt.Printf("Tick %d\n", count)

    if count >= 5 {
        break
    }
}
```

**Timer vs Ticker:**
```
Timer                           Ticker
─────                           ──────
Fires ONCE after duration       Fires REPEATEDLY at interval
time.NewTimer(2s)               time.NewTicker(500ms)
   │                               │
   ▼                               ▼
[wait 2s]                       [wait 500ms]
   │                               │
   ▼                               ▼
🔔 fire! → done                 🔔 tick! → [wait 500ms] → 🔔 tick! → ...
```

**Key points:**
- `ticker.C` - channel that receives at each interval
- `ticker.Stop()` - must call to prevent resource leak
- `for range ticker.C` - loop over ticks
- Use `break` to exit the loop when done

---

## 15. Buffered Channels for Burst Handling

Decouple producer and consumer speeds.

```go
// Unbuffered - producer blocks until consumer receives
jobs := make(chan Job)

// Buffered - producer can queue up to N items
jobs := make(chan Job, 10)
```

**Visual difference:**
```
Unbuffered (capacity 0):
Producer ──▶ [   ] ◀── Consumer
              ↑
         Must happen simultaneously


Buffered (capacity 3):
Producer ──▶ [J1][J2][J3] ──▶ Consumer
                   ↑
         Producer can "queue up" jobs
```

**Buffer size strategies:**
| Strategy | Buffer Size | Use Case |
|----------|-------------|----------|
| Match job count | `numJobs` | Producer finishes fast, workers process gradually |
| Match worker count | `numWorkers` | One pending job per worker |
| Fixed small buffer | `5-10` | General purpose, limits memory |

**Key points:**
- Buffered channels absorb speed differences
- Producer doesn't block until buffer is full
- Useful for handling bursts of work

---

## 16. Multiple Defer Statements

Multiple defers execute in LIFO order (Last In, First Out).

```go
func example() {
    defer fmt.Println("first defer")   // Runs 3rd
    defer fmt.Println("second defer")  // Runs 2nd
    defer fmt.Println("third defer")   // Runs 1st

    fmt.Println("doing work")
}

// Output:
// doing work
// third defer
// second defer
// first defer
```

**Practical example:**
```go
func (p *Producer) Start(jobs chan<- job.Job) {
    defer close(jobs)       // Runs 2nd - close channel

    ticker := time.NewTicker(p.tickInterval)
    defer ticker.Stop()     // Runs 1st - stop ticker

    // ... produce jobs ...
}
```

**Key points:**
- Defers stack up - last defer runs first
- Order matters for cleanup dependencies
- Each defer captures values at time of defer statement

---

## 17. Atomic Counters

Lock-free thread-safe counters using `sync/atomic`.

```go
import "sync/atomic"

type Metrics struct {
    processed atomic.Int64
    failed    atomic.Int64
}

// Increment (thread-safe)
m.processed.Add(1)

// Read (thread-safe)
count := m.processed.Load()
```

**Why atomic instead of regular int?**
```go
// Regular int - NOT SAFE (3 steps, can be interrupted)
m.processed++  // Read → Increment → Write

// Atomic - SAFE (single CPU instruction)
m.processed.Add(1)
```

**When to use atomic vs mutex:**
| Data Type | Use |
|-----------|-----|
| Simple counters (int64) | `atomic.Int64` |
| Maps, slices, structs | `sync.Mutex` |

**Key points:**
- `Add(n)` - atomically add to counter
- `Load()` - atomically read value
- `Store(n)` - atomically set value
- Faster than mutex for simple operations

---

## 18. Mutexes

Protect complex shared data with locks.

```go
import "sync"

type Metrics struct {
    mu         sync.Mutex
    failedJobs map[int]error
}

// Write with lock
func (m *Metrics) RecordFailure(jobID int, err error) {
    m.mu.Lock()
    m.failedJobs[jobID] = err
    m.mu.Unlock()
}

// Read with lock (use defer for safety)
func (m *Metrics) FailedJobs() map[int]error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Return a copy to prevent external modification
    result := make(map[int]error, len(m.failedJobs))
    for k, v := range m.failedJobs {
        result[k] = v
    }
    return result
}
```

**Visual:**
```
Goroutine A                         Goroutine B
───────────                         ───────────
mu.Lock() ✓ (acquired)
                                    mu.Lock() ⏳ (waiting...)
failedJobs[1] = err
mu.Unlock()
                                    mu.Lock() ✓ (acquired)
                                    failedJobs[2] = err
                                    mu.Unlock()
```

**Key points:**
- `Lock()` - acquire lock (blocks if held by another)
- `Unlock()` - release lock
- Use `defer m.mu.Unlock()` to ensure unlock on all paths
- Return copies of internal data to prevent races

---

## 19. Rate Limiting

Control throughput using token bucket pattern with Ticker.

```go
type Consumer struct {
    rateLimiter *time.Ticker
}

func New(rateLimit time.Duration) *Consumer {
    return &Consumer{
        rateLimiter: time.NewTicker(rateLimit),
    }
}

func (c *Consumer) Start(jobs <-chan Job) {
    defer c.rateLimiter.Stop()

    for job := range jobs {
        <-c.rateLimiter.C  // Wait for rate limit token
        c.process(job)
    }
}
```

**Token bucket concept:**
```
Every 100ms: +1 token (10 ops/second)

┌─────────────┐
│ ●  ●  ●  ●  │  ← 4 tokens available
└─────────────┘

To process: take a token (or wait for one)
```

**Rate calculation:**
```
rateLimit = 200ms → 5 jobs/sec per worker
3 workers × 5 jobs/sec = 15 jobs/sec max throughput
```

**Key points:**
- Each worker has its own rate limiter
- `<-rateLimiter.C` blocks until next tick
- Prevents overwhelming downstream services
- Always `Stop()` the ticker when done

---

## 20. Stateful Goroutines

Own state in a single goroutine, communicate via channels (no locks).

```go
type MetricsServer struct {
    updateCh chan MetricUpdate
    queryCh  chan StatsRequest
    done     chan struct{}
}

// Run as: go server.Run()
func (s *MetricsServer) Run() {
    // Internal state - only this goroutine touches these
    var processed, failed int64

    for {
        select {
        case update := <-s.updateCh:
            switch update.Type {
            case "success":
                processed++
            case "failure":
                failed++
            }
        case req := <-s.queryCh:
            req.ReplyCh <- StatsResponse{
                Processed: processed,
                Failed:    failed,
            }
        case <-s.done:
            return
        }
    }
}

// Helper methods hide channel details
func (s *MetricsServer) RecordSuccess() {
    s.updateCh <- MetricUpdate{Type: "success"}
}

func (s *MetricsServer) Stats() (processed, failed int64) {
    req := StatsRequest{ReplyCh: make(chan StatsResponse)}
    s.queryCh <- req
    resp := <-req.ReplyCh
    return resp.Processed, resp.Failed
}
```

**Mutex vs Stateful Goroutine:**
```
Mutex approach:
Worker 1 ──┐                    ┌── Lock
Worker 2 ──┼── shared data ◄───┼── Lock
Worker 3 ──┘                    └── Lock


Stateful Goroutine:
Worker 1 ──┐
Worker 2 ──┼──▶ [channel] ──▶ │OwnerGoroutine│ ──▶ owns data
Worker 3 ──┘                  │ (no locks)    │
```

**Request-Response pattern:**
```go
type StatsRequest struct {
    ReplyCh chan StatsResponse  // Caller provides reply channel
}

// Query: send request, wait for response
req := StatsRequest{ReplyCh: make(chan StatsResponse)}
s.queryCh <- req          // Send request
resp := <-req.ReplyCh     // Wait for response
```

**When to use which:**
| Aspect | Mutex | Stateful Goroutine |
|--------|-------|---------------------|
| State access | Any goroutine with lock | Single owner goroutine |
| Complexity | Simpler for small state | Better for state machines |
| Use case | Quick operations | Complex state logic |

**Key points:**
- Single goroutine owns all state (no races possible)
- Others communicate via channels
- Use buffered update channel to avoid blocking workers
- for-select loop is the core pattern

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Producer                                │
│                    (Tickers, Buffered Channel)                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      jobs channel                               │
│              (Buffered, Channel Directions)                     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
     ┌─────────┐      ┌─────────┐      ┌─────────┐
     │ Worker 1│      │ Worker 2│      │ Worker 3│
     │  select │      │  select │      │  select │
     │ +timeout│      │ +timeout│      │ +timeout│
     │ +rate   │      │ +rate   │      │ +rate   │
     │  limit  │      │  limit  │      │  limit  │
     └────┬────┘      └────┬────┘      └────┬────┘
          │                │                │
          ├────────────────┼────────────────┤
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    results channel                              │
│            (Non-blocking send, Buffered)                        │
└──────────────────────────┬──────────────────────────────────────┘
                           │
          ┌────────────────┴────────────────┐
          ▼                                 ▼
┌──────────────────────┐      ┌─────────────────────────────────┐
│    Main Collector    │      │       Metrics Server            │
│ (WaitGroup, Shutdown │      │    (Stateful Goroutine)         │
│       Timer)         │      │                                 │
└──────────────────────┘      │  ┌─────────────────────────┐    │
                              │  │ updateCh ──▶ for-select │    │
                              │  │ queryCh  ──▶   loop     │    │
                              │  │ done     ──▶ (owns:     │    │
                              │  │              processed, │    │
                              │  │              failed)    │    │
                              │  └─────────────────────────┘    │
                              └─────────────────────────────────┘

Workers also send metrics updates:
     ┌─────────┐
     │ Worker  │──── RecordSuccess() ────▶ │ Metrics Server │
     │         │──── RecordFailure() ────▶ │ (via channels) │
     └─────────┘
```

---

## Quick Reference

| Pattern | Use Case | Key Construct |
|---------|----------|---------------|
| Worker Pool | Parallel processing | `for range channel` in goroutines |
| Select | Multi-channel listen | `select { case ... }` |
| Timeout | Deadline for operations | `time.After` in select |
| Graceful Shutdown | Clean exit | `close(done)` + select |
| Non-blocking | Try without waiting | `select` with `default` |
| Broadcast | Signal all goroutines | `close(channel)` |
| Ticker | Periodic events | `time.NewTicker` + `for range ticker.C` |
| Buffered Channel | Burst handling | `make(chan T, capacity)` |
| Multiple Defer | Ordered cleanup | LIFO execution order |
| Atomic Counter | Lock-free counting | `atomic.Int64.Add()` / `.Load()` |
| Mutex | Protect complex data | `sync.Mutex.Lock()` / `.Unlock()` |
| Rate Limiting | Control throughput | Ticker as token bucket |
| Stateful Goroutine | State via channels | for-select loop owns data |
| Request-Response | Query stateful goroutine | Reply channel in request struct |

---

## Phase Summary

| Phase | Concepts Covered |
|-------|------------------|
| 1 | Goroutines, Channels, WaitGroups, Custom Errors |
| 2 | Worker Pools, Channel Directions, Range, Closing Channels |
| 3 | Select, Timeouts, Non-Blocking Ops, Timers |
| 4 | Tickers, Buffered Channels |
| 5 | Atomic Counters, Mutexes |
| 6 | Rate Limiting |
| 7 | Stateful Goroutines |

---

*Generated while building GoQueue - a concurrent job processing system in Go*
