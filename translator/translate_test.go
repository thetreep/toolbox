package translator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/require"
	"github.com/thetreep/toolbox/translator"
	"golang.org/x/text/language"
)

func TestTranslate(t *testing.T) {
	svc, _ := translator.New(bundlesFS, i18n.NewBundle(language.English))

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
			messageID: translator.HelloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world in french",
			langs:     []string{language.French.String()},
			messageID: translator.HelloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello world in a missing language",
			langs:     []string{language.Japanese.String()},
			messageID: translator.HelloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world in a specific language",
			langs:     []string{language.CanadianFrench.String()},
			messageID: translator.HelloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello world with no language",
			langs:     nil,
			messageID: translator.HelloWorld,
			want:      "Hello world!",
		},
		{
			name:      "hello world with a fallback language",
			langs:     []string{language.Japanese.String(), language.French.String()},
			messageID: translator.HelloWorld,
			want:      "Salut le monde !",
		},
		{
			name:      "hello you",
			langs:     []string{language.English.String()},
			messageID: translator.HelloYou,
			args:      "John",
			want:      "Hello John!",
		},
		{
			name:      "female gender",
			langs:     []string{language.French.String()},
			messageID: translator.SurroundedByFriends,
			args: map[string]string{
				"Gender": "female",
			},
			want: "entouré d'amie",
		},
		{
			name:      "male gender",
			langs:     []string{language.French.String()},
			messageID: translator.SurroundedByFriends,
			args: map[string]string{
				"Gender": "male",
			},
			want: "entouré d'ami",
		},
		{
			name:      "no binary gender",
			langs:     []string{language.French.String()},
			messageID: translator.SurroundedByFriends,
			args: map[string]string{
				"Gender": "neutral",
			},
			want: "entouré d'ami.e",
		},
		{
			name:      "no arguments",
			langs:     []string{language.French.String()},
			messageID: translator.SurroundedByFriends,
			want:      "entouré d'ami.e",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ctx := context.Background()
				ctx = translator.ContextWithLanguages(ctx, tt.langs...)

				translated, err := svc.Translate(ctx, tt.messageID, tt.args)
				if err != nil {
					t.Errorf(fmt.Errorf("translating: %w", err).Error())

					return
				}

				require.Equal(t, tt.want, translated)
			},
		)
	}
}
