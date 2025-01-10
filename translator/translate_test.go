package translator_test

import (
	"context"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/require"
	"github.com/thetreep/toolbox/translator"
	"golang.org/x/text/language"
)

// for testing purposes.
const (
	helloWorld = "helloWorld"
	helloYou   = "helloYou"
)

var (
	_ = &i18n.Message{
		ID:    helloWorld,
		Other: "Hello world!",
	}
	_ = &i18n.Message{
		ID:    helloYou,
		Other: "Hello {{.}}!",
	}
)

func TestTranslate(t *testing.T) {

	svc := translator.New(bundlesFS, i18n.NewBundle(language.English))

	tests := []struct {
		name      string
		langs     []string
		messageID string
		args      any
		want      string
	}{
		{
			name:      "hello world",
			langs:     []string{language.English.String()},
			messageID: helloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world in french",
			langs:     []string{language.French.String()},
			messageID: helloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello world in a missing language",
			langs:     []string{language.Japanese.String()},
			messageID: helloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world in a specific language",
			langs:     []string{language.CanadianFrench.String()},
			messageID: helloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello world with no language",
			langs:     nil,
			messageID: helloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world with a fallback language",
			langs:     []string{language.Japanese.String(), language.French.String()},
			messageID: helloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello you",
			langs:     []string{language.English.String()},
			messageID: helloYou,
			args:      "John",
			want:      "Hello John!",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ctx := context.Background()
				ctx = svc.ContextWithLanguages(ctx, tt.langs...)
				require.Equal(t, tt.want, svc.Translate(ctx, tt.messageID, tt.args))
			},
		)
	}
}
