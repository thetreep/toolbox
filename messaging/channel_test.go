package messaging_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thetreep/toolbox/messaging"
	"github.com/thetreep/toolbox/tests"
)

func Test_Messaging_Channel(t *testing.T) {
	tests.Setup(t, func(ctx context.Context) {
		dtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		channel := make(chan messaging.Message, 1)
		message := messaging.Message{}

		publisher := messaging.NewChannelPublisher(channel)
		err := publisher.Publish(ctx, message)
		assert.NoError(t, err, message)

		receiver := messaging.NewChannelReceiver(channel)
		assert.Equal(t, 1, len(channel))

		err = receiver.OnMessage(dtx,
			func(ctx context.Context,
				received messaging.Message) error {
				assert.Equal(t, message, received)

				return nil
			})
		assert.NoError(t, err, message)

		err = publisher.Close()
		assert.NoError(t, err, message)
	})
}
