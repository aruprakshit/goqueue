package producer

import (
	"fmt"

	"goqueue/job"
)

// Producer generates jobs and sends them to the jobs channel
type Producer struct {
	jobCount int
}

func New(jobCount int) *Producer {
	return &Producer{
		jobCount: jobCount,
	}
}

// Start begins producing jobs and sending them to the provided channel.
// It closes the channel when done to signal no more jobs will be sent.
func (p *Producer) Start(jobs chan<- job.Job) {
	defer close(jobs) // Signal consumers that no more jobs are coming

	for i := 1; i <= p.jobCount; i++ {
		j := job.NewJob(i, fmt.Sprintf("task-%d", i), fmt.Sprintf("payload-for-job-%d", i))
		fmt.Printf("[Producer] Created %v\n", j)
		jobs <- j // Send job to channel
	}

	fmt.Printf("[Producer] Finished producing %d jobs\n", p.jobCount)
}
