package producer

import (
	"goqueue/job"
	"testing"
)

func TestNew(t *testing.T) {
	p := New(10)

	if p.jobCount != 10 {
		t.Errorf("jobCount: got =  %d, expected = 10", p.jobCount)
	}
}

func TestStart(t *testing.T) {
	p := New(3)
	jobs := make(chan job.Job)

	go p.Start(jobs)

	var received []job.Job

	for j := range jobs {
		received = append(received, j)
	}

	if len(received) != 3 {
		t.Errorf("job count: got = %d, expected = 3", len(received))
	}

	if received[0].ID != 1 {
		t.Errorf("first job ID: got = %d, expected = 1", received[0].ID)
	}

	if received[2].ID != 3 {
		t.Errorf("last job ID: got = %d, expected = 3", received[2].ID)
	}
}
