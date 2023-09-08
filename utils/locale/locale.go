package locale

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed *.json
var localeFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
)

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	if _, err := bundle.LoadMessageFileFS(localeFS, "en.json"); err != nil {
		panic(err)
	}
	if _, err := bundle.LoadMessageFileFS(localeFS, "zh.json"); err != nil {
		panic(err)
	}
	localizer = i18n.NewLocalizer(bundle, "en")
}

func SetLocalizer(l *i18n.Localizer) {
	localizer = l
}

type Option func(c *i18n.LocalizeConfig)

func WithTplData(data interface{}) Option {
	return func(c *i18n.LocalizeConfig) {
		c.TemplateData = data
	}
}

func WithPluralCount(pluralCount interface{}) Option {
	return func(c *i18n.LocalizeConfig) {
		c.PluralCount = pluralCount
	}
}

func WithDefault(msg *i18n.Message) Option {
	return func(c *i18n.LocalizeConfig) {
		c.DefaultMessage = msg
	}
}

func T(msgID string, opts ...Option) (string, error) {
	c := &i18n.LocalizeConfig{MessageID: msgID}
	for _, opt := range opts {
		opt(c)
	}
	return localizer.Localize(c)
}

func MustT(msgID string, opts ...Option) string {
	msg, err := T(msgID, opts...)
	if err != nil {
		panic(err)
	}
	return msg
}
