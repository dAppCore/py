package i18n

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newI18nInterpreter registers the i18n module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newI18nInterpreter(t)
//	handle, err := caller.Call("core.i18n", "new")
func newI18nInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register i18n module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestI18n_Locales_Good adds and reports registered locales.
func TestI18n_Locales_Good(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.i18n", "add_locales", handle, "en", "fr"); callErr != nil {
		t.Fatalf("add_locales: %v", callErr)
	}

	value, callErr := caller.Call("core.i18n", "locales", handle)
	if callErr != nil {
		t.Fatalf("locales: %v", callErr)
	}
	list, ok := value.([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("locales: unexpected %#v", value)
	}
}

// TestI18n_TranslateWithoutTranslator_Good echoes the message id when no
// translator is registered.
func TestI18n_TranslateWithoutTranslator_Good(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.i18n", "translate", handle, "welcome.title")
	if callErr != nil {
		t.Fatalf("translate: %v", callErr)
	}
	if value != "welcome.title" {
		t.Fatalf("expected echoed message id, got %#v", value)
	}
}

// TestI18n_Language_Good reports the default and a set language.
func TestI18n_Language_Good(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.i18n", "language", handle)
	if callErr != nil {
		t.Fatalf("language default: %v", callErr)
	}
	if value != "en" {
		t.Fatalf("expected default language en, got %#v", value)
	}

	if _, callErr := caller.Call("core.i18n", "set_language", handle, "fr"); callErr != nil {
		t.Fatalf("set_language: %v", callErr)
	}
	value, callErr = caller.Call("core.i18n", "language", handle)
	if callErr != nil {
		t.Fatalf("language after set: %v", callErr)
	}
	if value != "fr" {
		t.Fatalf("expected fr, got %#v", value)
	}
}

// TestI18n_AvailableLanguages_Good reports the default language set when no
// translator is registered.
func TestI18n_AvailableLanguages_Good(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.i18n", "available_languages", handle)
	if callErr != nil {
		t.Fatalf("available_languages: %v", callErr)
	}
	if langs, ok := value.([]string); !ok || len(langs) != 1 || langs[0] != "en" {
		t.Fatalf("unexpected available languages %#v", value)
	}
}

// TestI18n_TranslatorNil_Good reports a nil translator handle.
func TestI18n_TranslatorNil_Good(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	value, callErr := caller.Call("core.i18n", "translator", handle)
	if callErr != nil {
		t.Fatalf("translator: %v", callErr)
	}
	if value != nil {
		t.Fatalf("expected nil translator, got %#v", value)
	}
}

// TestI18n_SetTranslator_Ugly rejects a value that is not a translator.
func TestI18n_SetTranslator_Ugly(t *core.T) {
	caller := newI18nInterpreter(t)

	handle, callErr := caller.Call("core.i18n", "new")
	if callErr != nil {
		t.Fatalf("new: %v", callErr)
	}

	if _, callErr := caller.Call("core.i18n", "set_translator", handle, "not-a-translator"); callErr == nil {
		t.Fatal("expected error for non-translator value")
	}
}

// TestI18n_WrongHandle_Ugly rejects a non-i18n handle argument.
func TestI18n_WrongHandle_Ugly(t *core.T) {
	caller := newI18nInterpreter(t)

	if _, callErr := caller.Call("core.i18n", "language", "not-a-handle"); callErr == nil {
		t.Fatal("expected error for non-i18n handle")
	}
}
