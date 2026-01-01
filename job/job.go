package job

import (
	"fmt"
	"time"
)

// Status represents the current state of a job
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Job represents a unit of work to be processed
type Job struct {
	ID        int
	Name      string
	Payload   string // data to process
	CreatedAt time.Time
}

func NewJob(id int, name, payload string) Job {
	return Job{
		ID:        id,
		Name:      name,
		Payload:   payload,
		CreatedAt: time.Now(),
	}
}

func (j Job) String() string {
	return fmt.Sprintf("Job{ID: %d, Name: %s}", j.ID, j.Name)
}

// Result represents the outcome of processing a job
type Result struct {
	JobID       int
	Status      Status
	Output      string
	Error       error
	ProcessedAt time.Time
	Duration    time.Duration
}

// creates a successful result
func NewResult(jobID int, output string, duration time.Duration) Result {
	return Result{
		JobID:       jobID,
		Status:      StatusCompleted,
		Output:      output,
		ProcessedAt: time.Now(),
		Duration:    duration,
	}
}

// creates a failed result
func NewErrorResult(jobID int, err error, duration time.Duration) Result {
	return Result{
		JobID:       jobID,
		Status:      StatusFailed,
		Error:       err,
		ProcessedAt: time.Now(),
		Duration:    duration,
	}
}

func (r Result) String() string {
	if r.Status == StatusFailed {
		return fmt.Sprintf("Result{JobID: %d, Status: %s, Error: %v}", r.JobID, r.Status, r.Error)
	}
	return fmt.Sprintf("Result{JobID: %d, Status: %s, Duration: %v}", r.JobID, r.Status, r.Duration)
}
