package translator

import (
	"context"
	"slices"

	"golang.org/x/text/language"
)

func ContextWithLanguages(ctx context.Context, languages ...string) context.Context {
	return context.WithValue(ctx, languagesCtxKey, languages)
}

func languagesFromContext(ctx context.Context) []string {
	langs, _ := ctx.Value(languagesCtxKey).([]string)
	return langs
}

func PreferredLanguageFromContext(ctx context.Context) language.Tag {
	supportedLangs := bundle.LanguageTags()
	lang, _ := language.MatchStrings(language.NewMatcher(supportedLangs), languagesFromContext(ctx)...)

	// for some reason the matcher can give variants of our languages…
	for !slices.Contains(supportedLangs, lang) {
		parent := lang.Parent()
		if parent == language.Und {
			break
		}
		lang = parent
	}

	return lang
}

type languagesCtxKeyType string

const languagesCtxKey languagesCtxKeyType = "languagesCtxKey"
