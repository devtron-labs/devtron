package connector

import (
	"context"
	"io"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/common-lib/constants"
	"go.uber.org/zap"
)

func newTestPump() PumpImpl {
	logger, _ := zap.NewDevelopment()
	runnable := async.NewAsyncRunnable(logger.Sugar(), constants.ServiceName("test"))
	return PumpImpl{logger: logger.Sugar(), asyncRunnable: runnable}
}

// blockingReader blocks on Read until Close is called, then returns io.EOF
type blockingReader struct {
	ch chan struct{}
}

func (b *blockingReader) Read(p []byte) (int, error) {
	<-b.ch
	return 0, io.EOF
}
func (b *blockingReader) Close() error {
	select {
	case <-b.ch:
		// already closed
	default:
		close(b.ch)
	}
	return nil
}

func TestStartK8sStreamWithHeartBeat_CtxCancelled_Returns(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		waitForStart bool // synchronise so cancel fires after goroutine entry
	}{
		{"cancel_before_goroutine_blocks", false},
		{"cancel_after_goroutine_entry", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pump := newTestPump()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			br := &blockingReader{ch: make(chan struct{})}
			w := httptest.NewRecorder()

			started := make(chan struct{})
			done := make(chan struct{})
			go func() {
				defer close(done)
				close(started)
				pump.StartK8sStreamWithHeartBeat(ctx, w, false, br, nil)
			}()

			if tc.waitForStart {
				<-started
			}
			cancel()

			select {
			case <-done:
				// success: function returned after ctx cancel without deadlock
			case <-time.After(3 * time.Second):
				t.Fatalf("StartK8sStreamWithHeartBeat did not return after ctx cancel (%s)", tc.name)
			}
		})
	}
}

// fakeStream returns a ReadCloser that emits fixed content then EOF.
type fakeStream struct {
	r    io.Reader
	done chan struct{}
}

func newFakeStream(content string) *fakeStream {
	return &fakeStream{r: strings.NewReader(content), done: make(chan struct{})}
}
func (f *fakeStream) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f *fakeStream) Close() error {
	select {
	case <-f.done:
	default:
		close(f.done)
	}
	return nil
}

func TestStartK8sStreamWithHeartBeat_MalformedLines_DoNotAbortStream(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		payload string
	}{
		{"blank_line_mid_stream", "2024-01-01T00:00:00Z hello world\n\n2024-01-01T00:00:01Z second line\n"},
		{"line_without_space", "2024-01-01T00:00:00Z hello\nnotimestamp\n2024-01-01T00:00:02Z after bad\n"},
		{"empty_stream", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pump := newTestPump()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stream := newFakeStream(tc.payload)
			w := httptest.NewRecorder()
			done := make(chan struct{})
			go func() {
				defer close(done)
				pump.StartK8sStreamWithHeartBeat(ctx, w, false, stream, nil)
			}()
			select {
			case <-done:
				// success: returned cleanly without deadlock
			case <-time.After(3 * time.Second):
				t.Fatalf("stream did not complete in time (%s)", tc.name)
			}
		})
	}
}

// sentinelFlusher reports a test error if Flush is called after markReturned().
type sentinelFlusher struct {
	*httptest.ResponseRecorder
	mu          sync.Mutex
	afterReturn bool
	t           *testing.T
}

func (sf *sentinelFlusher) markReturned() {
	sf.mu.Lock()
	sf.afterReturn = true
	sf.mu.Unlock()
}

func (sf *sentinelFlusher) Flush() {
	sf.mu.Lock()
	violation := sf.afterReturn
	sf.mu.Unlock()
	if violation {
		sf.t.Error("Flush called after StartK8sStreamWithHeartBeat returned — heartbeat goroutine leak")
	}
	sf.ResponseRecorder.Flush()
}

// TestHeartbeatGoroutineExitsBeforeFunctionReturns verifies that no Flush()
// call reaches the ResponseWriter after StartK8sStreamWithHeartBeat returns.
func TestHeartbeatGoroutineExitsBeforeFunctionReturns(t *testing.T) {
	t.Parallel()

	pump := newTestPump()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := newFakeStream("2024-01-01T00:00:00Z hello world\n")
	w := &sentinelFlusher{ResponseRecorder: httptest.NewRecorder(), t: t}

	done := make(chan struct{})
	go func() {
		defer close(done)
		pump.StartK8sStreamWithHeartBeat(ctx, w, false, stream, nil)
		w.markReturned() // any Flush after this line is a violation
	}()

	select {
	case <-done:
		// Give scheduler a moment: a leaked goroutine would call Flush here.
		time.Sleep(50 * time.Millisecond)
	case <-time.After(3 * time.Second):
		t.Fatal("StartK8sStreamWithHeartBeat did not return within 3s")
	}
}

// TestNoGoroutineLeakAfterStreamEOF verifies no goroutine is left running
// after the stream finishes normally.
func TestNoGoroutineLeakAfterStreamEOF(t *testing.T) {
	// Not parallel: runtime.NumGoroutine() counts all goroutines in the process;
	// parallel sibling tests contaminate the before/after samples.
	pump := newTestPump()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime.GC()
	before := runtime.NumGoroutine()

	stream := newFakeStream("2024-01-01T00:00:00Z line one\n")
	pump.StartK8sStreamWithHeartBeat(ctx, httptest.NewRecorder(), false, stream, nil)

	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	if after > before {
		t.Errorf("goroutine leak: %d before, %d after", before, after)
	}
}
