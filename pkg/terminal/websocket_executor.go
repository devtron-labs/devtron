package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// WebSocketExecutor handles terminal sessions over WebSocket
type WebSocketExecutor struct {
	config *rest.Config
	method string
	url    *url.URL
}

// NewWebSocketExecutor creates a new WebSocket-based executor for terminal sessions
func NewWebSocketExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	return &WebSocketExecutor{
		config: config,
		method: method,
		url:    url,
	}, nil
}

// Stream is deprecated. Please use StreamWithContext.
func (e *WebSocketExecutor) Stream(options remotecommand.StreamOptions) error {
	return e.StreamWithContext(context.Background(), options)
}

// StreamWithContext initiates the transport of the standard shell streams over WebSocket
func (e *WebSocketExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	// Create a WebSocket connection
	wsURL := e.url.String()
	wsURL = fmt.Sprintf("ws%s", wsURL[4:]) // Convert http(s) to ws(s)

	// Get transport from config
	transport, err := rest.TransportFor(e.config)
	if err != nil {
		return fmt.Errorf("error creating transport: %v", err)
	}

	// Create a WebSocket dialer with transport
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  transport.(*http.Transport).TLSClientConfig,
	}

	// Connect to WebSocket
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %v", err)
	}
	defer conn.Close()

	// Create channels for handling streams
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Handle stdin
	if options.Stdin != nil {
		go func() {
			defer close(doneChan)
			buf := make([]byte, 1024)
			for {
				n, err := options.Stdin.Read(buf)
				if err != nil {
					if err != io.EOF {
						errorChan <- fmt.Errorf("error reading from stdin: %v", err)
					}
					return
				}
				if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					errorChan <- fmt.Errorf("error writing to WebSocket: %v", err)
					return
				}
			}
		}()
	}

	// Handle stdout and stderr
	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					errorChan <- fmt.Errorf("error reading from WebSocket: %v", err)
				}
				return
			}

			if messageType == websocket.BinaryMessage {
				if options.Stdout != nil {
					if _, err := options.Stdout.Write(message); err != nil {
						errorChan <- fmt.Errorf("error writing to stdout: %v", err)
						return
					}
				}
			} else if messageType == websocket.TextMessage {
				// Handle control messages (like resize)
				var msg TerminalMessage
				if err := json.Unmarshal(message, &msg); err != nil {
					errorChan <- fmt.Errorf("error unmarshaling message: %v", err)
					return
				}

				switch msg.Op {
				case "resize":
					if options.TerminalSizeQueue != nil {
						options.TerminalSizeQueue.Next()
					}
				}
			}
		}
	}()

	// Wait for completion or error
	select {
	case err := <-errorChan:
		return err
	case <-doneChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
