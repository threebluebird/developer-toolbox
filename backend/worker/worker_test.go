package worker

import (
	"context"
	"errors"
	"testing"
)

func terminal(events <-chan Event) Event {
	var last Event
	for event := range events {
		last = event
	}
	return last
}

func TestWorkerCompletesWithProgress(t *testing.T) {
	w := New()
	_, events := w.Start(context.Background(), func(_ context.Context, report func(int, string)) (any, error) {
		report(50, "half")
		return "done", nil
	})
	last := terminal(events)
	if last.Status != Completed || last.Progress != 100 || last.Data != "done" {
		t.Fatalf("unexpected terminal event: %+v", last)
	}
}

func TestWorkerReportsErrors(t *testing.T) {
	w := New()
	_, events := w.Start(context.Background(), func(context.Context, func(int, string)) (any, error) {
		return nil, errors.New("boom")
	})
	if last := terminal(events); last.Status != Failed || last.Error != "boom" {
		t.Fatalf("unexpected terminal event: %+v", last)
	}
}

func TestWorkerCanCancel(t *testing.T) {
	w := New()
	started := make(chan struct{})
	id, events := w.Start(context.Background(), func(ctx context.Context, _ func(int, string)) (any, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	<-started
	if !w.Cancel(id) {
		t.Fatal("expected active job to be cancelled")
	}
	if last := terminal(events); last.Status != Cancelled {
		t.Fatalf("unexpected terminal event: %+v", last)
	}
}
