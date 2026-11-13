package runner

import (
	"context"
	"testing"
	"time"
)

// The first network chapter must become visible after fio even when a later
// diagnostic is still running. Collecting the whole group first regresses the
// v0.1.174 workflow and makes an otherwise completed chapter look hung.
func TestLegacyWorkflowPublishesCompletedChaptersBeforeSlowDiagnostic(t *testing.T) {
	for _, cancelAfterFirst := range []bool{false, true} {
		name := "complete"
		if cancelAfterFirst {
			name = "cancel"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			slowRelease := make(chan struct{})
			slowStarted := make(chan struct{})
			slowDone := make(chan struct{})
			emitted := make(chan string, 4)
			speedStarted := make(chan struct{}, 1)
			done := make(chan struct{})
			plan := legacyWorkflowPlan{
				hardware: []bufferedTask{{name: "disk", run: func(context.Context) string {
					return "fio finished\n"
				}}},
				independent: []bufferedTask{
					{name: "media", run: func(context.Context) string { return "media finished\n" }},
					{name: "routes", run: func(ctx context.Context) string {
						defer close(slowDone)
						close(slowStarted)
						select {
						case <-slowRelease:
							return "routes finished\n"
						case <-ctx.Done():
							return ""
						}
					}},
				},
				speed: func(context.Context) { speedStarted <- struct{}{} },
				emit:  func(value string) { emitted <- value },
			}
			go func() {
				runLegacyWorkflowPlan(ctx, plan)
				close(done)
			}()
			t.Cleanup(func() {
				cancel()
				<-done
				<-slowDone
			})
			for _, want := range []string{"fio finished\n", "media finished\n"} {
				select {
				case got := <-emitted:
					if got != want {
						t.Fatalf("chapter = %q, want %q", got, want)
					}
				case <-time.After(time.Second):
					t.Fatalf("completed chapter %q is waiting for a later diagnostic", want)
				}
			}
			<-slowStarted
			select {
			case <-speedStarted:
				t.Fatal("throughput overlapped the unfinished diagnostic")
			default:
			}
			if cancelAfterFirst {
				cancel()
			} else {
				close(slowRelease)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("workflow did not finish after release or cancellation")
			}
			if cancelAfterFirst {
				select {
				case <-speedStarted:
					t.Fatal("throughput started after cancellation")
				case value := <-emitted:
					t.Fatalf("chapter emitted after cancellation: %q", value)
				default:
				}
			} else {
				if got := <-emitted; got != "routes finished\n" {
					t.Fatalf("last diagnostic = %q", got)
				}
				select {
				case <-speedStarted:
				default:
					t.Fatal("throughput did not start after diagnostics completed")
				}
			}
		})
	}
}

func TestFullConcurrentWorkflowKeepsOutputBufferedUntilAllStagesFinish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan string, 3)
	finished := make(chan string, 3)
	releases := map[string]chan struct{}{
		"basic": make(chan struct{}), "disk": make(chan struct{}), "speed": make(chan struct{}),
	}
	task := func(name string) bufferedTask {
		return bufferedTask{name: name, run: func(ctx context.Context) string {
			started <- name
			select {
			case <-releases[name]:
			case <-ctx.Done():
			}
			finished <- name
			return name
		}}
	}
	emitted := make(chan string, 3)
	done := make(chan struct{})
	go func() {
		runLegacyWorkflowPlan(ctx, legacyWorkflowPlan{
			fullConcurrent: true, concurrentBasics: ptrBufferedTask(task("basic")),
			hardware: []bufferedTask{task("disk")}, concurrentSpeed: ptrBufferedTask(task("speed")),
			emit: func(value string) { emitted <- value },
		})
		close(done)
	}()
	t.Cleanup(func() { cancel(); <-done })
	for range 3 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("concurrent suite did not start every stage")
		}
	}
	close(releases["basic"])
	close(releases["disk"])
	<-finished
	<-finished
	select {
	case value := <-emitted:
		t.Fatalf("concurrent output escaped while speed was running: %q", value)
	case <-time.After(25 * time.Millisecond):
	}
	close(releases["speed"])
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("concurrent suite did not finish")
	}
	for _, want := range []string{"basic", "disk", "speed"} {
		if got := <-emitted; got != want {
			t.Fatalf("concurrent chapter = %q, want %q", got, want)
		}
	}
}
