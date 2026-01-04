package consumer

import (
	"fmt"
	"math/rand"
	"time"

	customerrors "goqueue/errors"
	"goqueue/job"
	"goqueue/metrics"
)

// Consumer processes jobs from the jobs channel
type Consumer struct {
	id          int
	maxRetries  int
	jobTimeout  time.Duration
	metrics     *metrics.Metrics
	rateLimiter *time.Ticker
}

// New creates a new Consumer with the given ID
func New(id int, m *metrics.Metrics, rateLimit time.Duration) *Consumer {
	return &Consumer{
		id:          id,
		maxRetries:  3,
		jobTimeout:  150 * time.Millisecond, // Jobs taking longer will timeout
		metrics:     m,
		rateLimiter: time.NewTicker(rateLimit),
	}
}

// Start begins consuming jobs from the provided channel.
// Results are sent to the results channel.
// The consumer stops when done is closed OR jobs channel is closed.
func (c *Consumer) Start(jobs <-chan job.Job, results chan<- job.Result, done <-chan struct{}) {
	defer c.rateLimiter.Stop() // stop the ticket when done
	for {
		select {
		case <-done:
			// Shutdown signal received
			fmt.Printf("[Consumer %d] Received shutdown signal\n", c.id)
			return
		case j, ok := <-jobs:
			if !ok {
				// Jobs channel closed
				fmt.Printf("[Consumer %d] No more jobs, shutting down\n", c.id)
				return
			}

			<-c.rateLimiter.C // wait for rate limit token

			result := c.process(j)
			// Non-blocking send: try to send, warn if channel is full
			select {
			case results <- result:
				// Sent successfully
			default:
				// Channel full - log warning but don't block
				fmt.Printf("[Consumer %d] WARNING: results channel full, dropping result for job %d\n", c.id, result.JobID)
			}
		}
	}
}

// process handles a single job and returns a Result
// Retries up to maxRetries times on failure, with timeout per attempt
func (c *Consumer) process(j job.Job) job.Result {
	startTime := time.Now()
	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		fmt.Printf("[Consumer %d] Processing %v (attempt %d/%d)\n", c.id, j, attempt, c.maxRetries)

		// Run work in goroutine so we can timeout
		resultCh := make(chan bool, 1) // true = success, false = random failure
		go func() {
			// Simulate work with random duration (50-200ms)
			workDuration := time.Duration(50+rand.Intn(150)) * time.Millisecond
			time.Sleep(workDuration)

			// Simulate random failures (20% chance)
			resultCh <- rand.Float32() >= 0.2 // true if success
		}()

		// Wait for work OR timeout
		select {
		case success := <-resultCh:
			if !success {
				lastErr = customerrors.NewJobError(j.ID, "processing failed", fmt.Errorf("random failure"))
				fmt.Printf("[Consumer %d] Attempt %d failed for %v, retrying...\n", c.id, attempt, j)
				continue // Retry
			}
			// Success!
			duration := time.Since(startTime)
			output := fmt.Sprintf("Processed payload: %s", j.Payload)
			fmt.Printf("[Consumer %d] Completed %v in %v\n", c.id, j, duration)
			c.metrics.RecordSuccess()
			return job.NewResult(j.ID, output, duration)

		case <-time.After(c.jobTimeout):
			lastErr = customerrors.NewTimeoutError(j.ID, c.jobTimeout.String())
			fmt.Printf("[Consumer %d] Attempt %d timed out for %v, retrying...\n", c.id, attempt, j)
			continue // Retry
		}
	}

	// All retries exhausted
	duration := time.Since(startTime)
	fmt.Printf("[Consumer %d] All retries exhausted for %v\n", c.id, j)
	c.metrics.RecordFailure(j.ID, lastErr)
	return job.NewErrorResult(j.ID, lastErr, duration)
}
