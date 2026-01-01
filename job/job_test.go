package job

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewJob(t *testing.T) {
	job := NewJob(11, "Bulk upload", "data:12,233,22")

	if job.ID != 11 {
		t.Errorf("got = %d, expected = 11", job.ID)
	}

	if job.Name != "Bulk upload" {
		t.Errorf("got = %q, expected = %q", job.Name, "Bulk upload")
	}

	if job.Payload != "data:12,233,22" {
		t.Errorf("got = %q, expected = %q", job.Payload, "data:12,233,22")
	}

	if job.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be set")
	}
}

func TestJob_String(t *testing.T) {
	job := NewJob(42, "Email task", "user@example.com")
	expected := "Job{ID: 42, Name: Email task}"
	got := fmt.Sprintf("%s", job)

	if got != expected {
		t.Errorf("got = %q, expected = %q", got, expected)
	}
}

func TestNewResult(t *testing.T) {
	duration := 30 * time.Millisecond
	result := NewResult(22, "processed successfully", duration)

	if result.JobID != 22 {
		t.Errorf("JobID: got = %d, expected = 22", result.JobID)
	}

	if result.Status != StatusCompleted {
		t.Errorf("Status: got = %q, expected = %q", result.Status, StatusCompleted)
	}

	if result.Output != "processed successfully" {
		t.Errorf("Output: got = %q, expected = %q", result.Output, "processed successfully")
	}

	if result.Duration != duration {
		t.Errorf("Duration: got = %v, expected = %v", result.Duration, duration)
	}

	if result.ProcessedAt.IsZero() {
		t.Errorf("ProcessedAt should be set")
	}

	if result.Error != nil {
		t.Errorf("Error should be blank for successfull result, got = %v", result.Error)
	}

}

func TestNewErrorResult(t *testing.T) {
	duration := 30 * time.Millisecond
	err := errors.New("connection timeout")
	result := NewErrorResult(22, err, duration)

	if result.JobID != 22 {
		t.Errorf("JobID: got = %d, expected = 22", result.JobID)
	}

	if result.Status != StatusFailed {
		t.Errorf("Status: got = %q, expected = %q", result.Status, StatusFailed)
	}

	if result.Output != "" {
		t.Errorf("Output should be empty for error result, got = %q", result.Output)
	}

	if result.Duration != duration {
		t.Errorf("Duration: got = %v, expected = %v", result.Duration, duration)
	}

	if result.ProcessedAt.IsZero() {
		t.Errorf("ProcessedAt should be set")
	}

	if result.Error != err {
		t.Errorf("Error: got = %v, expected = %v", result.Error, err)
	}

}

func TestResult_String(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected string
	}{
		{
			name:     "successful result",
			result:   NewResult(1, "done", 100*time.Millisecond),
			expected: "Result{JobID: 1, Status: completed, Duration: 100ms}",
		},
		{
			name:     "failed result",
			result:   NewErrorResult(1, errors.New("timeout"), 50*time.Millisecond),
			expected: "Result{JobID: 1, Status: failed, Error: timeout}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf("%s", tt.result)
			if got != tt.expected {
				t.Errorf("got = %q, expected = %q", got, tt.expected)
			}
		})
	}
}
