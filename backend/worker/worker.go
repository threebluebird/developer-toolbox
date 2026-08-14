// Package worker supplies cancellable asynchronous execution without imposing
// a scheduling system on tools.
package worker

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Status string

const (
	Running   Status = "running"
	Progress  Status = "progress"
	Completed Status = "completed"
	Cancelled Status = "cancelled"
	Failed    Status = "error"
)

type Event struct {
	JobID    string `json:"jobId"`
	Status   Status `json:"status"`
	Progress int    `json:"progress"`
	Message  string `json:"message,omitempty"`
	Data     any    `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Task func(ctx context.Context, report func(progress int, message string)) (any, error)

type Worker struct {
	mu     sync.Mutex
	cancel map[string]context.CancelFunc
}

func New() *Worker { return &Worker{cancel: make(map[string]context.CancelFunc)} }

// Start begins a task and returns its stable ID and event stream. The channel
// is closed after exactly one terminal event.
func (w *Worker) Start(parent context.Context, task Task) (string, <-chan Event) {
	if parent == nil {
		parent = context.Background()
	}
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(parent)
	events := make(chan Event, 8)
	w.mu.Lock()
	w.cancel[id] = cancel
	w.mu.Unlock()

	go w.run(ctx, id, task, events)
	return id, events
}

func (w *Worker) Cancel(jobID string) bool {
	w.mu.Lock()
	cancel, ok := w.cancel[jobID]
	w.mu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

func (w *Worker) run(ctx context.Context, id string, task Task, events chan Event) {
	defer close(events)
	defer func() {
		w.mu.Lock()
		delete(w.cancel, id)
		w.mu.Unlock()
	}()
	events <- Event{JobID: id, Status: Running}
	report := func(progress int, message string) {
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
		select {
		case events <- Event{JobID: id, Status: Progress, Progress: progress, Message: message}:
		case <-ctx.Done():
		}
	}
	data, err := task(ctx, report)
	if ctx.Err() != nil {
		events <- Event{JobID: id, Status: Cancelled, Error: ctx.Err().Error()}
		return
	}
	if err != nil {
		events <- Event{JobID: id, Status: Failed, Error: err.Error()}
		return
	}
	events <- Event{JobID: id, Status: Completed, Progress: 100, Data: data}
}
