package messaging

import (
	"context"
	"fmt"
	"time"

	pubsub "cloud.google.com/go/pubsub"
	"github.com/cockroachdb/errors"
	logging "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// GooglePublisher publishes messages on google pubsub.
type GooglePublisher struct {
	topic  *pubsub.Topic
	client *pubsub.Client
}

// NewGooglePublisher returns a publisher using google pubsub.
func NewGooglePublisher(
	ctx context.Context, projectID, topicName, credFile string) (Publisher, error) {
	client, err := pubsub.NewClient(
		ctx,
		projectID,
		option.WithCredentialsFile(credFile))
	if err != nil {
		return nil, errors.WithStack(
			errors.Wrapf(
				err,
				"an error occurred while setting a google client for project %s",
				projectID),
		)
	}

	topic := client.Topic(topicName)
	if topic == nil {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, errors.WithStack(
				errors.Wrapf(
					err,
					"an error occurred while creating topic %s on project %s because it doesnt exist",
					topicName, projectID),
			)
		}
	}

	doExist, err := topic.Exists(ctx)
	if err != nil {
		return nil, errors.WithStack(
			errors.Wrapf(
				err,
				"an error occurred while checking: %s for project %s",
				projectID, topicName),
		)
	}

	if !doExist {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, errors.WithStack(
				errors.Wrapf(
					err,
					"an error occurred while creating topic %s on project %s because it doesnt exist",
					topicName, projectID),
			)
		}
	}

	return &GooglePublisher{
		topic:  topic,
		client: client,
	}, nil
}

// Close releases resources used by this Publisher.
func (e *GooglePublisher) Close() error {
	err := e.client.Close()
	if err != nil {
		return errors.WithStack(
			errors.Wrap(err, "an error occurred while closing the underlying pubsub client"),
		)
	}

	return nil
}

// Publish send message to pubsub.
func (e *GooglePublisher) Publish(ctx context.Context, message Message) error {
	result := e.topic.Publish(ctx, &pubsub.Message{
		Data: message.Data,
	})

	if result == nil {
		return errors.WithStack(
			errors.New("received an empty response from google"))
	}

	_, err := result.Get(ctx)
	if err != nil {
		return errors.WithStack(
			errors.Wrap(err,
				"an error occurred while pushing a message to pubsub"))
	}

	return nil
}

// GoogleReceiver returns a Receiver using google pubsub.
type GoogleReceiver struct {
	subscription *pubsub.Subscription
	client       *pubsub.Client
}

// NewGoogleReceiver returns a Receiver that listens to Google pubsub.
func NewGoogleReceiver(
	ctx context.Context, projectID, subscriptionName, creadFile string) (Receiver, error) {
	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(creadFile))
	if err != nil {
		return nil, errors.WithStack(
			errors.Wrapf(
				err,
				"an error occurred while setting a google client for project %s",
				projectID))
	}

	subscription := client.Subscription(subscriptionName)
	if subscription == nil {
		return nil, errors.WithStack(
			errors.Wrapf(
				err,
				"an error occurred while getting subscription: %s for project %s",
				subscriptionName, projectID),
		)
	}

	doExist, err := subscription.Exists(ctx)
	if err != nil {
		return nil, errors.WithStack(
			errors.Wrapf(
				err,
				"an error occurred while checking subscription: %s for project %s",
				subscriptionName, projectID),
		)
	}

	if !doExist {
		return nil, errors.WithStack(
			errors.Newf(
				"subscription %s doesnt exist for project %s",
				subscriptionName, projectID),
		)
	}

	return &GoogleReceiver{
		subscription: subscription,
		client:       client,
	}, nil
}

// OnMessage registers handler for a given underlying subscription.
func (e *GoogleReceiver) OnMessage(ctx context.Context,
	handler func(ctx context.Context, message Message) error) error {
	if e.subscription == nil {
		return errors.WithStack(errors.New("internal subscription is nil"))
	}

	logging.Info("starting google pubsub receiver")

	e.subscription.ReceiveSettings.MaxExtension = 30 * time.Minute
	e.subscription.ReceiveSettings.MaxOutstandingMessages = 2
	e.subscription.ReceiveSettings.MaxOutstandingBytes = 10e6

	err := e.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if msg == nil {
			logging.Error("received empty message from google pubsub")
			return
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		attributes := make(map[string]interface{})
		for key, value := range msg.Attributes {
			attributes[fmt.Sprintf("%v", key)] = value
		}

		logger := logging.WithFields(attributes).WithField("id", msg.ID)

		logger.Infof(
			"received on subscription name: %s id: %s",
			e.subscription.ID(), e.subscription.String())

		err := handler(ctx, Message{Data: msg.Data})
		if err != nil {
			logger.WithError(err).Error("an error occurred while handling message")
			msg.Nack()
			return
		}

		logger.Info("successfully handled message")
		msg.Ack()
	})
	if err != nil {
		return errors.WithStack(
			errors.Wrapf(err,
				"an error occurred while receiving from %s", e.subscription.ID()))
	}

	err = e.client.Close()
	if err != nil {
		return errors.WithStack(
			errors.Wrap(err,
				"an error occurred while closing the underlying pubsub client "))
	}

	return nil
}
