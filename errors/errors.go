package errors

import "fmt"

// JobError represents a custom error for job processing failures
type JobError struct {
  JobID   int
  Message string
  Err     error // underlying error (if any)
}

func (e *JobError) Error() string {
  if e.Err != nil {
    return fmt.Sprintf("job %d failed: %s: %v", e.JobID, e.Message, e.Err)
  }
  return fmt.Sprintf("job %d failed: %s", e.JobID, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *JobError) Unwrap() error {
  return e.Err
}

// NewJobError creates a new JobError
func NewJobError(jobID int, message string, err error) *JobError {
  return &JobError{
    JobID:   jobID,
    Message: message,
    Err:     err,
  }
}

// ValidationError represents invalid job input
type ValidationError struct {
  Field   string
  Message string
}

func (e *ValidationError) Error() string {
  return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
  return &ValidationError{
    Field:   field,
    Message: message,
  }
}

// TimeoutError represents a job that exceeded its time limit
type TimeoutError struct {
  JobID    int
  Duration string
}

func (e *TimeoutError) Error() string {
  return fmt.Sprintf("job %d timed out after %s", e.JobID, e.Duration)
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(jobID int, duration string) *TimeoutError {
  return &TimeoutError{
    JobID:    jobID,
    Duration: duration,
  }
}
