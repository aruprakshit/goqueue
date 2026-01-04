package consumer

import (
	"fmt"
	"math/rand"
	"time"

	customerrors "goqueue/errors"
	"goqueue/job"
)

// Consumer processes jobs from the jobs channel
type Consumer struct {
	id         int
	maxRetries int
}

// New creates a new Consumer with the given ID
func New(id int) *Consumer {
	return &Consumer{id: id, maxRetries: 3}
}

// Start begins consuming jobs from the provided channel.
// Results are sent to the results channel.
// The consumer runs until the jobs channel is closed.
func (c *Consumer) Start(jobs <-chan job.Job, results chan<- job.Result) {
	for j := range jobs {
		result := c.process(j)
		results <- result
	}
	fmt.Printf("[Consumer %d] No more jobs, shutting down\n", c.id)
}

// process handles a single job and returns a Result
// Retries up to maxRetries times on failure
func (c *Consumer) process(j job.Job) job.Result {
	startTime := time.Now()
	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		fmt.Printf("[Consumer %d] Processing %v (attempt %d/%d)\n", c.id, j, attempt, c.maxRetries)

		// Simulate work with random duration (50-200ms)
		workDuration := time.Duration(50+rand.Intn(150)) * time.Millisecond
		time.Sleep(workDuration)

		// Simulate random failures (20% chance)
		if rand.Float32() < 0.2 {
			lastErr = customerrors.NewJobError(j.ID, "processing failed", fmt.Errorf("random failure"))
			fmt.Printf("[Consumer %d] Attempt %d failed for %v, retrying...\n", c.id, attempt, j)
			continue // Retry
		}

		// Success!
		duration := time.Since(startTime)
		output := fmt.Sprintf("Processed payload: %s", j.Payload)
		fmt.Printf("[Consumer %d] Completed %v in %v\n", c.id, j, duration)
		return job.NewResult(j.ID, output, duration)
	}

	// All retries exhausted
	duration := time.Since(startTime)
	fmt.Printf("[Consumer %d] All retries exhausted for %v\n", c.id, j)
	return job.NewErrorResult(j.ID, lastErr, duration)
}
