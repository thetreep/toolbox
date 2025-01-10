package translator

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/thetreep/toolbox/logger"
	"gopkg.in/yaml.v3"
)

// //go:generate ./gen.sh
// //go:embed *.yaml
// var bundlesFS embed.FS

// //nolint:gochecknoglobals // we don't need to mock translations, we can use a global
// var bundle = i18n.NewBundle(language.English)

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

func (svc TranslationService) Translate(ctx context.Context, messageID string, args any) string {
	return svc.TranslateWithPlural(ctx, messageID, nil, args)
}

func (svc TranslationService) TranslateWithPlural(
	ctx context.Context, messageID string, count interface{}, args any,
) string {
	langs := languagesFromContext(ctx)
	localizer := i18n.NewLocalizer(svc.bundle, langs...)

	localizedMessage, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: args,
			PluralCount:  count,
		},
	)
	if err != nil {
		logger.Error(
			ctx,
			fmt.Errorf("translation error: %w", err),
			slog.Any("messageID", messageID),
			slog.Any("pluralCount", count),
			slog.Any("args", args),
		)

		return fmt.Sprintf("translation_error(%s)", messageID)
	}

	return localizedMessage
}
