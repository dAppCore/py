package i18n

import (
	"fmt"

	core "dappco.re/go/core"
	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

type translator interface {
	Translate(messageID string, args ...any) core.Result
	SetLanguage(lang string) error
	Language() string
	AvailableLanguages() []string
}

type handle struct {
	locales    []any
	locale     string
	translator translator
}

// Register exposes Core i18n helpers.
//
//	i18n.Register(interpreter)
func Register(interpreter *runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.i18n",
		Documentation: "Locale and translation helpers for CorePy",
		Functions: map[string]runtime.Function{
			"new":                 newI18n,
			"add_locales":         addLocales,
			"locales":             locales,
			"set_translator":      setTranslator,
			"translator":          translatorValue,
			"translate":           translate,
			"set_language":        setLanguage,
			"language":            language,
			"available_languages": availableLanguages,
		},
	})
}

func newI18n(arguments ...any) (any, error) {
	return &handle{}, nil
}

func addLocales(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.add_locales")
	if err != nil {
		return nil, err
	}
	item.locales = append(item.locales, arguments[1:]...)
	return item, nil
}

func locales(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.locales")
	if err != nil {
		return nil, err
	}
	result := make([]any, len(item.locales))
	copy(result, item.locales)
	return result, nil
}

func setTranslator(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.set_translator")
	if err != nil {
		return nil, err
	}
	if len(arguments) < 2 || arguments[1] == nil {
		item.translator = nil
		return item, nil
	}
	translatorValue, ok := arguments[1].(translator)
	if !ok {
		return nil, fmt.Errorf("core.i18n.set_translator expected translator, got %T", arguments[1])
	}
	item.translator = translatorValue
	if item.locale != "" {
		if err := item.translator.SetLanguage(item.locale); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func translatorValue(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.translator")
	if err != nil {
		return nil, err
	}
	if item.translator == nil {
		return nil, nil
	}
	return item.translator, nil
}

func translate(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.translate")
	if err != nil {
		return nil, err
	}
	messageID, err := typemap.ExpectString(arguments, 1, "core.i18n.translate")
	if err != nil {
		return nil, err
	}
	if item.translator == nil {
		return messageID, nil
	}
	return typemap.ResultValue(item.translator.Translate(messageID, arguments[2:]...), "core.i18n.translate")
}

func setLanguage(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.set_language")
	if err != nil {
		return nil, err
	}
	lang, err := typemap.ExpectString(arguments, 1, "core.i18n.set_language")
	if err != nil {
		return nil, err
	}
	if lang == "" {
		return item, nil
	}
	item.locale = lang
	if item.translator != nil {
		if err := item.translator.SetLanguage(lang); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func language(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.language")
	if err != nil {
		return nil, err
	}
	if item.locale != "" {
		return item.locale, nil
	}
	if item.translator != nil && item.translator.Language() != "" {
		return item.translator.Language(), nil
	}
	return "en", nil
}

func availableLanguages(arguments ...any) (any, error) {
	item, err := expectHandle(arguments, 0, "core.i18n.available_languages")
	if err != nil {
		return nil, err
	}
	if item.translator == nil {
		return []string{"en"}, nil
	}
	return item.translator.AvailableLanguages(), nil
}

func expectHandle(arguments []any, index int, functionName string) (*handle, error) {
	if index >= len(arguments) {
		return nil, fmt.Errorf("%s expected argument %d", functionName, index)
	}
	value, ok := arguments[index].(*handle)
	if !ok {
		return nil, fmt.Errorf("%s expected i18n handle, got %T", functionName, arguments[index])
	}
	return value, nil
}
