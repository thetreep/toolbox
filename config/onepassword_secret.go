package config

import (
	"context"
	"strings"

	"github.com/1password/onepassword-sdk-go"
)

const (
	OPSecretPrefix           = "op://"
	OPServiceAccountTokenKey = "OP_SERVICE_ACCOUNT_TOKEN"
)

// ResolveSecretFromOP takes a configuration value and, if it starts with OPSecretPrefix, replaces it with the secret value.
func ResolveSecretFromOP(ctx context.Context, value string, options ...resolveSecretFromOPOption) (string, error) {
	// just return the value if it doesn't start with the prefix
	if !strings.HasPrefix(value, OPSecretPrefix) {
		return value, nil
	}

	// token to use for authentication
	token, err := getToken(options...)

	// parse options
	integrationName := "unkown"
	integrationVersion := "unkown"
	for _, option := range options {
		switch {
		case option.integrationName != "":
			integrationName = option.integrationName
		case option.integrationVersion != "":
			integrationVersion = option.integrationVersion
		}
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

type resolveSecretFromOPOption struct {
	token              string
	integrationName    string
	integrationVersion string
}

// WithOPToken sets the token to use for authentication.
func WithOPToken(token string) resolveSecretFromOPOption {
	return resolveSecretFromOPOption{token: token}
}

// WithIntegrationName sets the name of the integration e.g. consumer app name.
func WithIntegrationName(name string) resolveSecretFromOPOption {
	return resolveSecretFromOPOption{integrationName: name}
}

// WithIntegrationVersion sets the version of the integration e.g. consumer app version.
func WithIntegrationVersion(version string) resolveSecretFromOPOption {
	return resolveSecretFromOPOption{integrationVersion: version}
}

// getToken to use for authentication.
// Defaults to OP_SERVICE_ACCOUNT_TOKEN env var.
// Otherwise, needs to be set with WithOPToken.
func getToken(options ...resolveSecretFromOPOption) (string, error) {
	var token string
	for _, option := range options {
		if option.token != "" {
			token = option.token
		}
	}

	if token == "" {
		var err error
		token, err = GetEnvOrError(OPServiceAccountTokenKey)
		if err != nil {
			return "", err
		}
	}

	return token, nil
}
