package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ravenpair/cli/internal/ports"
)

// Client is the WebSocket adapter that implements ports.WSClient.
type Client struct{}

// New returns a new WebSocket Client adapter.
func New() *Client {
	return &Client{}
}

// Dial connects to wsURL, sets the provided headers, and calls onMessage for
// each message received. It blocks until ctx is cancelled or the connection is
// closed by the server.
func (c *Client) Dial(ctx context.Context, wsURL string, headers map[string]string, onMessage ports.MessageHandler) error {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(ctx, wsURL, h)
	if err != nil {
		return fmt.Errorf("WebSocket dial: %w", err)
	}
	defer conn.Close()

	errCh := make(chan error, 1)
	go func() {
		for {
			msgType, data, readErr := conn.ReadMessage()
			if readErr != nil {
				errCh <- readErr
				return
			}
			onMessage(msgType, data)
		}
	}()

	select {
	case <-ctx.Done():
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		_ = conn.WriteMessage(websocket.CloseMessage, msg)
		return nil
	case err := <-errCh:
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return nil
		}
		return err
	}
}
