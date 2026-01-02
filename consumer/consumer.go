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
	id int
}

// New creates a new Consumer with the given ID
func New(id int) *Consumer {
	return &Consumer{id: id}
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
func (c *Consumer) process(j job.Job) job.Result {
	startTime := time.Now()
	fmt.Printf("[Consumer %d] Processing %v\n", c.id, j)

	// Simulate work with random duration (50-200ms)
	workDuration := time.Duration(50+rand.Intn(150)) * time.Millisecond
	time.Sleep(workDuration)

	duration := time.Since(startTime)

	// Simulate random failures (20% chance)
	if rand.Float32() < 0.2 {
		err := customerrors.NewJobError(j.ID, "processing failed", fmt.Errorf("random failure"))
		fmt.Printf("[Consumer %d] Failed %v: %v\n", c.id, j, err)
		return job.NewErrorResult(j.ID, err, duration)
	}

	output := fmt.Sprintf("Processed payload: %s", j.Payload)
	fmt.Printf("[Consumer %d] Completed %v in %v\n", c.id, j, duration)
	return job.NewResult(j.ID, output, duration)
}
