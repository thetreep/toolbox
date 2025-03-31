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

func (svc TranslationService) PreferredLanguageFromContext(
	ctx context.Context,
) language.Tag {
	supportedLangs := svc.bundle.LanguageTags()
	lang, _ := language.MatchStrings(
		language.NewMatcher(supportedLangs),
		languagesFromContext(ctx)...,
	)

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

// LangFromCtx returns the first valid language found in the context.
// Defaults to EN.
func LangFromCtx(ctx context.Context) language.Tag {
	langs := languagesFromContext(ctx)
	for _, langStr := range langs {
		tag, err := language.Parse(langStr)
		if err == nil {
			return tag
		}
	}

	return language.English
}

type languagesCtxKeyType string

const languagesCtxKey languagesCtxKeyType = "languagesCtxKey"
