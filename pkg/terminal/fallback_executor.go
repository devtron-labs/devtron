package terminal

import (
	"context"
	"fmt"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// FallbackExecutor tries WebSocket first and falls back to SPDY if needed
type FallbackExecutor struct {
	wsExecutor   remotecommand.Executor
	spdyExecutor remotecommand.Executor
}

// NewFallbackExecutor creates a new executor that tries WebSocket first and falls back to SPDY
func NewFallbackExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	// Create WebSocket executor
	wsExecutor, err := NewWebSocketExecutor(config, method, url)
	if err != nil {
		return nil, fmt.Errorf("error creating WebSocket executor: %v", err)
	}

	// Create SPDY executor as fallback
	spdyExecutor, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return nil, fmt.Errorf("error creating SPDY executor: %v", err)
	}

	return &FallbackExecutor{
		wsExecutor:   wsExecutor,
		spdyExecutor: spdyExecutor,
	}, nil
}

// Stream is deprecated. Please use StreamWithContext.
func (f *FallbackExecutor) Stream(options remotecommand.StreamOptions) error {
	return f.StreamWithContext(context.Background(), options)
}

// StreamWithContext tries WebSocket first and falls back to SPDY if needed
func (f *FallbackExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	// Try WebSocket first
	err := f.wsExecutor.StreamWithContext(ctx, options)
	if err == nil {
		return nil
	}

	// If WebSocket fails, try SPDY
	return f.spdyExecutor.StreamWithContext(ctx, options)
}
