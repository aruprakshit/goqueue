package consumer

import (
	"goqueue/job"
	"testing"
)

func TestNew(t *testing.T) {
	c := New(1)

	if c.id != 1 {
		t.Errorf("id: got = %d, expected = 1", c.id)
	}
}

func TestStart(t *testing.T) {
	c := New(1)
	jobs := make(chan job.Job)
	results := make(chan job.Result)

	go c.Start(jobs, results)

	go func() {
		jobs <- job.NewJob(1, "task-1", "payload-1")
		jobs <- job.NewJob(2, "task-2", "payload-2")
		jobs <- job.NewJob(3, "task-3", "payload-3")
	}()

	var received []job.Result
	for i := 0; i < 3; i++ {
		received = append(received, <-results)
	}

	if len(received) != 3 {
		t.Fatalf("result count: got = %d, expected = 3", len(received))
	}
}
