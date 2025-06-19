package terminal

import (
	"context"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// WebSocketExecutor wraps the Kubernetes WebSocket executor
type WebSocketExecutor struct {
	executor remotecommand.Executor
}

// NewWebSocketExecutor creates a new WebSocket-based executor for terminal sessions
func NewWebSocketExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	// Use the Kubernetes WebSocket executor directly
	executor, err := remotecommand.NewWebSocketExecutor(config, method, url.String())
	if err != nil {
		return nil, err
	}

	return &WebSocketExecutor{
		executor: executor,
	}, nil
}

// Stream is deprecated. Please use StreamWithContext.
func (e *WebSocketExecutor) Stream(options remotecommand.StreamOptions) error {
	return e.executor.Stream(options)
}

// StreamWithContext delegates to the underlying Kubernetes WebSocket executor
func (e *WebSocketExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	return e.executor.StreamWithContext(ctx, options)
}
