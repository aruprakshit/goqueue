package metrics

// MetricUpdate represents an update to send to the metrics server
type MetricUpdate struct {
	Type  string // "success" or "failure"
	JobID int
	Err   error
}

// StatsRequest is used to query current stats
type StatsRequest struct {
	ReplyCh chan StatsResponse
}

// StatsResponse contains the current metrics
type StatsResponse struct {
	Processed int64
	Failed    int64
}

// MetricServer manages metrics state in a dedicated goroutine
type MetricsServer struct {
	updateCh chan MetricUpdate
	queryCh  chan StatsRequest
	done     chan struct{}
}

// creates a new metric server
func NewServer() *MetricsServer {
	return &MetricsServer{
		updateCh: make(chan MetricUpdate, 100), // Buffered to avoid blocking workers
		queryCh:  make(chan StatsRequest),
		done:     make(chan struct{}),
	}
}

// sends a success update to the server
func (s *MetricsServer) RecordSuccess() {
	s.updateCh <- MetricUpdate{Type: "success"}
}

// sends a failure update to the server
func (s *MetricsServer) RecordFailure(jobID int, err error) {
	s.updateCh <- MetricUpdate{Type: "failure", JobID: jobID, Err: err}
}

// Stats queries the server for current stats (blocks until response)
func (s *MetricsServer) Stats() (processed, failed int64) {
	req := StatsRequest{ReplyCh: make(chan StatsResponse)}
	s.queryCh <- req
	resp := <-req.ReplyCh
	return resp.Processed, resp.Failed
}

// shuts down the metrics server
func (s *MetricsServer) Stop() {
	close(s.done)
}

// Run starts the metrics server
func (s *MetricsServer) Run() {
	// Internal state - only this goroutine touches these
	var processed, failed int64

	for {
		select {
		case update := <-s.updateCh:
			// Handle metric updates
			switch update.Type {
			case "success":
				processed++
			case "failure":
				failed++
			}
		case req := <-s.queryCh:
			// Respond to stats queries
			req.ReplyCh <- StatsResponse{
				Processed: processed,
				Failed:    failed,
			}
		case <-s.done:
			return
		}
	}
}
