package translator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestPreferredLanguageFromContext(t *testing.T) {
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
				ctx := ContextWithLanguages(context.Background(), tt.header)
				require.Equal(t, tt.want.String(), PreferredLanguageFromContext(ctx).String())
			},
		)
	}
}
