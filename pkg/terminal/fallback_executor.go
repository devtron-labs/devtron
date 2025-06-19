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
	// Create WebSocket executor
	wsExecutor, err := remotecommand.NewWebSocketExecutor(config, method, url.String())
	if err != nil {
		log.Printf("Warning: Failed to create WebSocket executor: %v", err)
		// Continue with SPDY only
		spdyExecutor, err := remotecommand.NewSPDYExecutor(config, method, url)
		if err != nil {
			return nil, fmt.Errorf("failed to create SPDY executor: %v", err)
		}
		return spdyExecutor, nil
	}

	// Create SPDY executor as fallback
	spdyExecutor, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		log.Printf("Warning: Failed to create SPDY executor: %v", err)
		// Use WebSocket only
		return wsExecutor, nil
	}

	// Use Kubernetes built-in fallback executor
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
