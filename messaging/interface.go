package messaging

import (
	"context"
)

// Message specifies
type Message struct {
	Data []byte
}

// Publisher specifies how to publish a message
// on any type of messaging infrastructure.
type Publisher interface {
	Publish(ctx context.Context, message Message) error
	Close() error
}

// Receiver specifies what happens when a message
// received from the messaging infrastructure.
type Receiver interface {
	OnMessage(ctx context.Context,
		handler func(ctx context.Context, message Message) error) error
}
