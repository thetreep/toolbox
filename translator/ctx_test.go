package translator_test

import (
	"context"
	"embed"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/require"
	"github.com/thetreep/toolbox/translator"
	"golang.org/x/text/language"
)

//go:embed *.yaml
var bundlesFS embed.FS

func TestPreferredLanguageFromContext(t *testing.T) {

	svc := translator.New(bundlesFS, i18n.NewBundle(language.English))

	tests := []struct {
		header string
		want   language.Tag
	}{
		{
			header: "en-US,en;q=0.5",
			want:   language.English,
		},
		{
			header: "fr-CA,en-US,en;q=0.5",
			want:   language.French,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.header, func(t *testing.T) {
				ctx := svc.ContextWithLanguages(context.Background(), tt.header)
				require.Equal(t, tt.want.String(), svc.PreferredLanguageFromContext(ctx).String())
			},
		)
	}
}
