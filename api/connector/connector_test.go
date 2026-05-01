package connector

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestPump() PumpImpl {
	logger, _ := zap.NewDevelopment()
	return PumpImpl{logger: logger.Sugar()}
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
