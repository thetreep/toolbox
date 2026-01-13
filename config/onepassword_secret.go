package config

import (
	"context"
	"strings"

	"github.com/1password/onepassword-sdk-go"
)

const (
	OPSecretPrefix = "op://"
	//nolint:gosec // not credentials
	OPServiceAccountTokenKey = "OP_SERVICE_ACCOUNT_TOKEN"
)

// ResolveSecretFromOP takes a configuration value and, if it starts with OPSecretPrefix, replaces it with the secret value.
func ResolveSecretFromOP(ctx context.Context, value string, options ...OPOption) (string, error) {
	// just return the value if it doesn't start with the prefix
	if !strings.HasPrefix(value, OPSecretPrefix) {
		return value, nil
	}

	opts := &opOptions{}
	for _, option := range options {
		option(opts)
	}

	// token to use for authentication
	token, err := getToken(opts)
	if err != nil {
		return "", err
	}

	// parse options
	integrationName := "unkown"
	if opts.integrationName != "" {
		integrationName = opts.integrationName
	}
	integrationVersion := "unkown"
	if opts.integrationVersion != "" {
		integrationVersion = opts.integrationVersion
	}

	// fetch the secret
	client, err := onepassword.NewClient(
		ctx,
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(integrationName, integrationVersion),
	)
	if err != nil {
		return value, err
	}
	return client.Secrets().Resolve(ctx, value)
}

type opOptions struct {
	token              string
	integrationName    string
	integrationVersion string
}

type OPOption func(o *opOptions)

// WithOPToken sets the token to use for authentication.
func WithOPToken(token string) OPOption {
	return func(o *opOptions) {
		o.token = token
	}
}

// WithIntegrationName sets the name of the integration e.g. consumer app name.
func WithIntegrationName(name string) OPOption {
	return func(o *opOptions) {
		o.integrationName = name
	}
}

// WithIntegrationVersion sets the version of the integration e.g. consumer app version.
func WithIntegrationVersion(version string) OPOption {
	return func(o *opOptions) {
		o.integrationVersion = version
	}
}

// getToken to use for authentication.
// Defaults to OP_SERVICE_ACCOUNT_TOKEN env var.
// Otherwise, needs to be set with WithOPToken.
func getToken(options *opOptions) (string, error) {
	if options.token != "" {
		return options.token, nil
	}

	return GetEnvOrError(OPServiceAccountTokenKey)
}
