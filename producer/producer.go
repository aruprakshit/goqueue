package producer

import (
	"fmt"
	"goqueue/job"
	"time"
)

// Producer generates jobs and sends them to the jobs channel
type Producer struct {
	jobCount     int
	tickInterval time.Duration
}

func New(jobCount int, tickInterval time.Duration) *Producer {
	return &Producer{
		jobCount:     jobCount,
		tickInterval: tickInterval,
	}
}

// Start begins producing jobs and sending them to the provided channel.
// It closes the channel when done to signal no more jobs will be sent.
func (p *Producer) Start(jobs chan<- job.Job) {
	defer close(jobs) // Signal consumers that no more jobs are coming

	ticker := time.NewTicker(p.tickInterval)
	defer ticker.Stop()

	count := 0

	for range ticker.C {
		count++
		j := job.NewJob(count, fmt.Sprintf("task-%d", count), fmt.Sprintf("payload-for-job-%d", count))
		fmt.Printf("[Producer] Created %v\n", j)
		jobs <- j // Send job to channel

		if count >= p.jobCount {
			break
		}
	}

	fmt.Printf("[Producer] Finished producing %d jobs\n", p.jobCount)
}
