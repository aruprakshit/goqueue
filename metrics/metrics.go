package metrics

import (
	"sync"
	"sync/atomic"
)

// Metrics tracks job processing statistics using atomic counters.
//
// Why atomic.Int64 instead of regular int?
//   - Regular int: processed++ is 3 steps (Read → Increment → Write), can be interrupted
//   - Atomic: processed.Add(1) is a single CPU instruction, cannot be interrupted
type Metrics struct {
	processed  atomic.Int64
	failed     atomic.Int64
	mu         sync.Mutex
	failedJobs map[int]error
}

func New() *Metrics {
	return &Metrics{
		failedJobs: make(map[int]error),
	}
}

// RecordSuccess increments the processed counter (thread-safe)
func (m *Metrics) RecordSuccess() {
	m.processed.Add(1)
}

// RecordFailure increments the failed counter (thread-safe)
func (m *Metrics) RecordFailure(jobID int, err error) {
	m.failed.Add(1)
	m.mu.Lock()
	m.failedJobs[jobID] = err
	m.mu.Unlock()
}

// Stats returns current counts (thread-safe)
func (m *Metrics) Stats() (processed, failed int64) {
	return m.processed.Load(), m.failed.Load()
}

func (m *Metrics) FailedJobs() map[int]error {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[int]error, len(m.failedJobs))
	for k, v := range m.failedJobs {
		result[k] = v
	}

	return result
}
