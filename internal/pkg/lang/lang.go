package lang

import (
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	defaultLanguage = "en"
)

var (
	bundle     *i18n.Bundle
	initBundle sync.Once
	languages  = []string{"en"}
	basePath   = "./resources/locales"
)

func SetBasePath(path string) {
	basePath = path
}

func SetSupportedLanguages(langs string) {
	// remove all the spaces in a langs string.
	replaceLangs := strings.ReplaceAll(langs, " ", "")

	languages = strings.Split(replaceLangs, ",")
}

func GetLocalizer(lang string) *i18n.Localizer {
	initBundle.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

		for _, lang := range languages {
			path := filepath.Join(basePath, lang+".yml")

			_, err := bundle.LoadMessageFile(path)
			if err != nil {
				slog.Error("could not load message file",
					slog.String("error", err.Error()),
					slog.String("path", path),
				)
			}
		}
	})

	return i18n.NewLocalizer(bundle, lang)
}

type Localizable struct {
	Message     string
	MessageID   string
	MessageVars map[string]interface{}
}

func (l Localizable) Localize(lang string) string {
	if lang == "" {
		lang = defaultLanguage
	}

	localized, _ := GetLocalizer(lang).Localize(&i18n.LocalizeConfig{
		MessageID:    l.MessageID,
		TemplateData: l.MessageVars,
		DefaultMessage: &i18n.Message{
			ID:    l.MessageID,
			Other: l.Message,
		},
	})

	return localized
}
