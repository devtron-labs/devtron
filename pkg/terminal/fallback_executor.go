package terminal

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// FallbackExecutor tries WebSocket first and falls back to SPDY if needed
type FallbackExecutor struct {
	executor remotecommand.Executor
}

// NewFallbackExecutor creates a new executor that tries WebSocket first and falls back to SPDY
func NewFallbackExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	var wsExecutor, spdyExecutor remotecommand.Executor
	var wsErr, spdyErr error

	// Try to create WebSocket executor
	wsExecutor, wsErr = remotecommand.NewWebSocketExecutor(config, method, url.String())
	if wsErr != nil {
		log.Printf("Warning: Failed to create WebSocket executor: %v", wsErr)
	}

	// Try to create SPDY executor
	spdyExecutor, spdyErr = remotecommand.NewSPDYExecutor(config, method, url)
	if spdyErr != nil {
		log.Printf("Warning: Failed to create SPDY executor: %v", spdyErr)
	}

	// Handle different scenarios
	if wsErr != nil && spdyErr != nil {
		// Both failed
		return nil, fmt.Errorf("failed to create any executor: WebSocket error: %v, SPDY error: %v", wsErr, spdyErr)
	}

	if wsErr != nil {
		// Only WebSocket failed, use SPDY
		return spdyExecutor, nil
	}

	if spdyErr != nil {
		// Only SPDY failed, use WebSocket
		return wsExecutor, nil
	}

	// Both succeeded, create fallback executor
	fallbackExecutor, err := remotecommand.NewFallbackExecutor(wsExecutor, spdyExecutor, func(err error) bool {
		// Fall back to SPDY if WebSocket fails due to connection issues
		log.Printf("WebSocket failed, falling back to SPDY: %v", err)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback executor: %v", err)
	}

	return &FallbackExecutor{
		executor: fallbackExecutor,
	}, nil
}

// Stream is deprecated. Please use StreamWithContext.
func (e *FallbackExecutor) Stream(options remotecommand.StreamOptions) error {
	return e.executor.Stream(options)
}

// StreamWithContext delegates to the underlying fallback executor
func (e *FallbackExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	return e.executor.StreamWithContext(ctx, options)
}
