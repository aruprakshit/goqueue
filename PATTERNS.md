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

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Producer                                │
│                    (Rate Limiting, Tickers)                     │
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
     └────┬────┘      └────┬────┘      └────┬────┘
          │                │                │
          └────────────────┼────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    results channel                              │
│            (Non-blocking send, Buffered)                        │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Main Collector                             │
│              (WaitGroup, Shutdown Timer)                        │
└─────────────────────────────────────────────────────────────────┘
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

---

*Generated while building GoQueue - a concurrent job processing system in Go*
