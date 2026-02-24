package ports

import "context"

// MessageHandler is called for every message received from the WebSocket server.
type MessageHandler func(msgType int, data []byte)

// WSClient is the outgoing port for a persistent WebSocket connection.
type WSClient interface {
	// Dial connects to wsURL, forwarding headers, and calls onMessage for each
	// message until ctx is cancelled or the connection is closed.
	Dial(ctx context.Context, wsURL string, headers map[string]string, onMessage MessageHandler) error
}
