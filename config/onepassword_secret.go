package config

import (
	"context"

	"github.com/1password/onepassword-sdk-go"
)

const (
	OPSecretPrefix           = "op://"
	OPServiceAccountTokenKey = "OP_SERVICE_ACCOUNT_TOKEN"
)

// ResolveSecretFromOP takes a configuration value and, if it starts with OPSecretPrefix, replaces it with the secret value.
func ResolveSecretFromOP(ctx context.Context, value string, options ...resolveSecretFromOPOption) (string, error) {
	token, err := getToken(options...)

	client, err := onepassword.NewClient(
		ctx,
		onepassword.WithServiceAccountToken(token),
	)
	if err != nil {
		return value, err
	}
	return client.Secrets().Resolve(ctx, value)
}

type resolveSecretFromOPOption struct {
	token string
}

// WithOPToken sets the token to use for authentication.
func WithOPToken(token string) resolveSecretFromOPOption {
	return resolveSecretFromOPOption{token: token}
}

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
