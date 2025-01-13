package translator

import (
	"context"
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gopkg.in/yaml.v3"
)

type TranslationService struct {
	bundlesFS embed.FS
	bundle    *i18n.Bundle
}

func New(bundlesFS embed.FS, bundle *i18n.Bundle) TranslationService {
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	_, err := bundle.LoadMessageFileFS(bundlesFS, "active.en.yaml")
	if err != nil {
		panic(err)
	}

	_, err = bundle.LoadMessageFileFS(bundlesFS, "active.fr.yaml")
	if err != nil {
		panic(err)
	}

	return TranslationService{
		bundlesFS: bundlesFS,
		bundle:    bundle,
	}
}

func (svc TranslationService) Translate(
	ctx context.Context,
	messageID string,
	args any,
) (string, error) {
	return svc.TranslateWithPlural(ctx, messageID, nil, args)
}

func (svc TranslationService) TranslateWithPlural(
	ctx context.Context, messageID string, count interface{}, args any,
) (string, error) {
	langs := languagesFromContext(ctx)
	localizer := i18n.NewLocalizer(svc.bundle, langs...)

	return localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: args,
			PluralCount:  count,
		},
	)
}
