package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestJobError_Error(t *testing.T) {
	// Arrange
	underlyingErr := fmt.Errorf("connection refused")
	jobErr := NewJobError(42, "processing failed", underlyingErr)

	// Act
	got := jobErr.Error()

	// Assert
	expected := "job 42 failed: processing failed: connection refused"
	if expected != got {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestJobError_Error_WithoutUnderlyingError(t *testing.T) {
	// Arrange
	jobErr := NewJobError(2, "processing failed", nil)

	// Act
	got := jobErr.Error()

	// Assert
	expected := "job 2 failed: processing failed"
	if expected != got {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestJobError_WrapsUnderlyingError(t *testing.T) {
	underlyingErr := fmt.Errorf("connection refused")
	jobErr := NewJobError(11, "save failed", underlyingErr)

	if !errors.Is(jobErr, underlyingErr) {
		t.Errorf("expected JobError to wrap the the underlying error")
	}
}

func TestJobError_CanBeExtractedWithErrorsAs(t *testing.T) {
	jobErr := NewJobError(11, "processing failed", nil)
	wrappedErr := fmt.Errorf("operation failed: %w", jobErr)

	var extracted *JobError
	if !errors.As(wrappedErr, &extracted) {
		t.Fatal("expected to extract JobError from wrapped error")
	}

	if extracted.JobID != 11 {
		t.Errorf("got JobID %d, want 11", extracted.JobID)
	}
}

func TestValidationError_Error(t *testing.T) {
	validationErr := NewValidationError("ID", "can't be blank")

	got := validationErr.Error()
	expected := "validation error on field 'ID': can't be blank"

	if expected != got {
		t.Errorf("got = %q, expected = %q", got, expected)
	}
}

func TestValidationError_CanBeExtractedWithErrorsAs(t *testing.T) {
	validationErr := NewValidationError("ID", "can't be blank")
	wrappedErr := fmt.Errorf("operation failed: %w", validationErr)

	var extracted *ValidationError
	if !errors.As(wrappedErr, &extracted) {
		t.Fatal("expected to extract ValidationError from wrapped error")
	}

	if extracted.Field != "ID" {
		t.Errorf("got Field %s, want 'ID'", extracted.Field)
	}
}

func TestTimeoutError_Error(t *testing.T) {
	timeoutErr := NewTimeoutError(11, "4 seconds")

	got := timeoutErr.Error()
	expected := "job 11 timed out after 4 seconds"

	if expected != got {
		t.Errorf("got = %q, expected = %q", got, expected)
	}
}

func TestTimeoutError_CanBeExtractedWithErrorsAs(t *testing.T) {
	timeoutErr := NewTimeoutError(11, "4 seconds")
	wrappedErr := fmt.Errorf("operation failed: %w", timeoutErr)

	var extracted *TimeoutError
	if !errors.As(wrappedErr, &extracted) {
		t.Fatal("expected to extract TimeoutError from wrapped error")
	}

	if extracted.JobID != 11 {
		t.Errorf("got job id %d, want 11", extracted.JobID)
	}
}
