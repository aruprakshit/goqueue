package main

import (
	"fmt"
	"sync"
	"time"

	"goqueue/consumer"
	"goqueue/job"
	"goqueue/producer"
)

func main() {
	const (
		numJobs         = 5                    // Number of jobs to process
		numWorkers      = 3                    // Number of concurrent workers
		shutdownTimeout = 2 * time.Second // Max time before forced shutdown
	)

	// Create channels
	jobs := make(chan job.Job)
	results := make(chan job.Result, numJobs) // Buffered to prevent blocking
	// Signal channels use struct{} instead of bool because:
	// 1. Zero memory allocation (struct{} is 0 bytes, bool is 1 byte)
	// 2. Self-documenting: signals "only the event matters, not the value"
	// 3. Idiomatic Go pattern for broadcast signals via close()
	done := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		p := producer.New(numJobs)
		p.Start(jobs) // This will close the jobs channel when done
	}()

	// Start worker pool
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c := consumer.New(workerID)
			c.Start(jobs, results, done)
		}(i)
	}

	// Graceful shutdown timer
	shutdownTimer := time.NewTimer(shutdownTimeout)
	allDone := make(chan struct{})

	// Signal when all workers finish normally
	go func() {
		wg.Wait()
		close(allDone)
	}()

	// Wait for either: all done OR timeout
	go func() {
		select {
		case <-allDone:
			// Normal completion - stop the timer
			shutdownTimer.Stop()
			fmt.Println("[Main] All workers finished normally")
		case <-shutdownTimer.C:
			// Timeout - signal workers to stop
			fmt.Println("[Main] Shutdown timeout reached, stopping workers...")
			close(done)
		}
		close(results)
	}()

	fmt.Println("\n=== Results ===")
	var completed, failed int
	for result := range results {
		if result.Status == job.StatusCompleted {
			completed++
			fmt.Printf("✓ Job %d: %s (took %v)\n", result.JobID, result.Output, result.Duration)
		} else {
			failed++
			fmt.Printf("✗ Job %d: %v\n", result.JobID, result.Error)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Total: %d | Completed: %d | Failed: %d\n", completed+failed, completed, failed)
}
