package api

import (
	"context"
	"testing"
)

func TestProgressObserverReceivesSectionTransitions(t *testing.T) {
	var events []ProgressEvent
	ctx := WithProgressObserver(context.Background(), func(event ProgressEvent) {
		events = append(events, event)
	})
	report := collectComponentStep(ctx, "cpu", func() ComponentReport {
		return ComponentReport{Name: "cputest", Status: ReportStatusOK}
	})
	if report.Status != ReportStatusOK || len(events) != 2 {
		t.Fatalf("unexpected report/events: %#v %#v", report, events)
	}
	if events[0].Section != "cpu" || events[0].Phase != ProgressStarted || events[1].Phase != ProgressCompleted || events[1].Status != ReportStatusOK {
		t.Fatalf("unexpected progress events: %#v", events)
	}
}

func TestProgressObserverPanicDoesNotAbortRun(t *testing.T) {
	ctx := WithProgressObserver(context.Background(), func(ProgressEvent) { panic("observer failure") })
	report := collectComponentStep(ctx, "cpu", func() ComponentReport {
		return ComponentReport{Name: "cputest", Status: ReportStatusOK}
	})
	if report.Status != ReportStatusOK {
		t.Fatalf("observer changed report: %#v", report)
	}
}
