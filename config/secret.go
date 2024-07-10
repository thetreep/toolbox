package config

import (
	"context"
	"strings"

	"braces.dev/errtrace"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const GCPSecretPrefix = "gcp-secret://"

// ResolveSecretFromGCP takes a configuration value and, if it starts with GCPSecretPrefix, replaces it with the secret value.
//
// The value after GCPSecretPrefix must be a resource name in one of these two formats:
// `projects/*/secrets/*/versions/*` or `projects/*/locations/*/secrets/*/versions/*`
//
// Available options: WithGCPSecretClient
func ResolveSecretFromGCP(ctx context.Context, value string, options ...ResolveSecretFromGCPOption) (string, error) {
	if !strings.HasPrefix(value, GCPSecretPrefix) {
		return value, nil
	}
	client, err := resolveGCPSecretOptions(ctx, options)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	secretName := strings.TrimPrefix(value, GCPSecretPrefix)

	result, err := client.AccessSecretVersion(
		ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: secretName,
		},
	)
	if err != nil {
		return "", errtrace.Errorf("accessing GCP secret: %w", err)
	}

	return string(result.Payload.Data), nil
}

func resolveGCPSecretOptions(ctx context.Context, options []ResolveSecretFromGCPOption) (*secretmanager.Client, error) {
	var client *secretmanager.Client
	for _, option := range options {
		client = option(client)
	}
	if client == nil {
		var err error
		client, err = secretmanager.NewClient(ctx)
		if err != nil {
			return nil, errtrace.Errorf("creating default secretmanager client: %w", err)
		}
	}
	return client, nil
}

type ResolveSecretFromGCPOption func(*secretmanager.Client) *secretmanager.Client

func WithGCPSecretClient(client *secretmanager.Client) ResolveSecretFromGCPOption {
	return func(_ *secretmanager.Client) *secretmanager.Client {
		return client
	}
}
