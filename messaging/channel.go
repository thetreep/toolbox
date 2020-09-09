package messaging

import (
	"context"

	"github.com/cockroachdb/errors"
	logging "github.com/sirupsen/logrus"
)

// PublishToChannel implements the Publisher interface by using a channel.
type PublishToChannel struct {
	channel chan Message
}

// Publish push a message to a channel.
func (e *PublishToChannel) Publish(ctx context.Context, message Message) error {
	if e.channel == nil {
		return errors.New("internal channel is empty or not initialized")
	}

	select {
	case e.channel <- message:
		return nil
	case <-ctx.Done():
		return errors.New("could not publish before timeout")
	}
}

// Close is here just to satify the Publisher interface.
func (e *PublishToChannel) Close() error {
	close(e.channel)
	return nil
}

// NewChannelPublisher returns a Publisher that uses a channel.
func NewChannelPublisher(channel chan Message) Publisher {
	return &PublishToChannel{
		channel: channel,
	}
}

// NewChannelReceiver returns Receiver that uses a channel.
func NewChannelReceiver(channel chan Message) Receiver {
	return &ReceiveFromChannel{
		channel: channel,
	}
}

// ReceiveFromChannel implements the Receiver interface by using a channel.
type ReceiveFromChannel struct {
	channel chan Message
}

// OnMessage for receiving message from a channel.
func (e *ReceiveFromChannel) OnMessage(ctx context.Context,
	handler func(context.Context, Message) error) error {
	if e.channel == nil {
		return errors.New("internal channel is empty or not initialized")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-e.channel:
			if !ok {
				return nil
			}

			go func() {
				err := handler(ctx, message)
				if err != nil {
					logging.WithError(err).Error("messaging handler returned a error")
				}
			}()
		}
	}
}
