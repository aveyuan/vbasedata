package vbasedata

import (
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewPond_Defaults(t *testing.T) {
	c := &PondConfig{}
	p := NewPond(c, discardLogger())
	defer p.Stop()

	if c.MinWorkers != 2 {
		t.Errorf("MinWorkers default = %d, want 2", c.MinWorkers)
	}
	if c.MaxWorkers != 10 {
		t.Errorf("MaxWorkers default = %d, want 10", c.MaxWorkers)
	}
	if c.StopAndWait != 5 {
		t.Errorf("StopAndWait default = %d, want 5", c.StopAndWait)
	}
	if p.GetPond() == nil {
		t.Error("GetPond returned nil")
	}
}

func TestPond_SubmitRunsTasks(t *testing.T) {
	p := NewPond(&PondConfig{MaxWorkers: 4, MinWorkers: 2}, discardLogger())

	const n = 100
	var counter int64
	for i := 0; i < n; i++ {
		p.Submit(func() {
			atomic.AddInt64(&counter, 1)
		})
	}

	// Stop waits for queued/running tasks to finish.
	p.Stop()

	if got := atomic.LoadInt64(&counter); got != n {
		t.Errorf("executed %d tasks, want %d", got, n)
	}
}
