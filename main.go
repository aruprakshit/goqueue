package main

import (
	"fmt"
	"sync"

	"goqueue/consumer"
	"goqueue/job"
	"goqueue/producer"
)

func main() {
	const (
		numJobs    = 5 // Number of jobs to process
		numWorkers = 3 // Number of concurrent workers
	)

	// Create channels
	// jobs channel: producer sends jobs, consumer receives
	// results channel: consumer sends results, main collects
	jobs := make(chan job.Job)
	results := make(chan job.Result, numJobs) // Buffered to prevent blocking

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
			c.Start(jobs, results)
		}(i)
	}

	// Wait for producer and consumer to finish, then close results
	go func() {
		wg.Wait()
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
