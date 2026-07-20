package api

import (
	"context"
	"time"
)

type ProgressPhase string

const (
	ProgressStarted   ProgressPhase = "started"
	ProgressCompleted ProgressPhase = "completed"
)

// ProgressEvent reports a real structured section transition. Observers run
// synchronously and should return quickly.
type ProgressEvent struct {
	Section string        `json:"section"`
	Phase   ProgressPhase `json:"phase"`
	Status  ReportStatus  `json:"status,omitempty"`
	Reason  string        `json:"reason,omitempty"`
	At      time.Time     `json:"at"`
}

type ProgressObserver func(ProgressEvent)

type progressObserverContextKey struct{}

// WithProgressObserver returns a context that receives section transitions
// from RunAllTestsContext and the local structured component adapters.
func WithProgressObserver(parent context.Context, observer ProgressObserver) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if observer == nil {
		return parent
	}
	return context.WithValue(parent, progressObserverContextKey{}, observer)
}

func emitProgress(ctx context.Context, event ProgressEvent) {
	if ctx == nil {
		return
	}
	observer, _ := ctx.Value(progressObserverContextKey{}).(ProgressObserver)
	if observer == nil {
		return
	}
	if event.At.IsZero() {
		event.At = time.Now()
	}
	func() {
		defer func() { _ = recover() }()
		observer(event)
	}()
}

func progressStarted(ctx context.Context, section string) {
	emitProgress(ctx, ProgressEvent{Section: section, Phase: ProgressStarted})
}

func progressCompleted(ctx context.Context, section string, status ReportStatus, reason string) {
	emitProgress(ctx, ProgressEvent{Section: section, Phase: ProgressCompleted, Status: status, Reason: reason})
}

func collectComponentStep(ctx context.Context, section string, run func() ComponentReport) ComponentReport {
	progressStarted(ctx, section)
	report := run()
	progressCompleted(ctx, section, report.Status, report.Reason)
	return report
}
