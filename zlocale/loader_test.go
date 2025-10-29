package zlocale

import (
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewLoader(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	tt.NotNil(i18n)

	loader := NewLoader(i18n)
	tt.NotNil(loader)
	tt.Equal(i18n, loader.i18n)
}

func TestLoaderLoadTranslationsFromMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	loader := NewLoader(i18n)

	translations := map[string]string{
		"hello":     "Hello",
		"goodbye":   "Goodbye",
		"user.name": "John Doe",
	}

	err := loader.LoadTranslationsFromMap("en", "English", translations)
	tt.NoError(err)
	tt.EqualTrue(i18n.HasLanguage("en"))
	tt.Equal(i18n.T("hello"), "Hello")
	tt.Equal(i18n.T("user.name"), "John Doe")

	err = loader.LoadTranslationsFromMap("fr", "French", nil)
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "cannot be nil"))

	additionalTranslations := map[string]string{
		"welcome": "Welcome!",
		"thanks":  "Thank you",
	}

	err = loader.LoadTranslationsFromMap("en", "English", additionalTranslations)
	tt.NoError(err)

	tt.Equal(i18n.T("welcome"), "Welcome!")
	tt.Equal(i18n.T("thanks"), "Thank you")
}

func TestLoaderAddCustomTranslation(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	loader := NewLoader(i18n)

	baseTranslations := map[string]string{
		"hello":   "Hello",
		"goodbye": "Goodbye",
	}
	err := loader.LoadTranslationsFromMap("en", "English", baseTranslations)
	tt.NoError(err)

	err = loader.AddCustomTranslation("en", "welcome", "Welcome!")
	tt.NoError(err)
	tt.Equal(i18n.T("welcome"), "Welcome!")

	tt.Equal(i18n.T("welcome"), "Welcome!")

	err = loader.AddCustomTranslation("zh-CN", "hello", "你好")
	tt.NoError(err)
	tt.EqualTrue(i18n.HasLanguage("zh-CN"))

	err = loader.AddCustomTranslation("ja", "hello", "こんにちは")
	tt.NoError(err)

	languages := i18n.GetLoadedLanguages()
	tt.Equal("日本語", languages["ja"])
}

func TestLoaderGetAvailableLanguages(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	loader := NewLoader(i18n)

	languages := loader.GetAvailableLanguages()
	tt.Equal(0, len(languages))

	translations1 := map[string]string{"hello": "Hello"}
	translations2 := map[string]string{"hello": "Hola"}

	err := loader.LoadTranslationsFromMap("en", "English", translations1)
	tt.NoError(err)

	err = loader.LoadTranslationsFromMap("es", "Spanish", translations2)
	tt.NoError(err)

	languages = loader.GetAvailableLanguages()
	tt.Equal(2, len(languages))

	hasEn := false
	hasEs := false
	for _, code := range languages {
		if code == "en" {
			hasEn = true
		}
		if code == "es" {
			hasEs = true
		}
	}

	tt.EqualTrue(hasEn)
	tt.EqualTrue(hasEs)
}

func TestLoaderConcurrentOperations(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	loader := NewLoader(i18n)

	go func() {
		translations := map[string]string{"test1": "Test 1"}
		loader.LoadTranslationsFromMap("en", "English", translations)
	}()

	go func() {
		translations := map[string]string{"test2": "Test 2"}
		loader.LoadTranslationsFromMap("en", "English", translations)
	}()

	loader.LoadTranslationsFromMap("en", "English", map[string]string{"test3": "Test 3"})

	go func() {
		loader.AddCustomTranslation("en", "custom1", "Custom 1")
	}()

	go func() {
		loader.AddCustomTranslation("en", "custom2", "Custom 2")
	}()

	loader.AddCustomTranslation("en", "custom3", "Custom 3")

	tt.EqualTrue(i18n.HasLanguage("en"))

	translations := []string{"test1", "test2", "test3", "custom1", "custom2", "custom3"}
	foundCount := 0
	for _, key := range translations {
		if i18n.T(key) != key {
			foundCount++
		}
	}

	tt.EqualTrue(foundCount >= 1)
}

func TestGlobalLoaderFunctions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	translations := map[string]string{
		"global.hello": "Hello World",
		"global.bye":   "Goodbye",
	}

	err := LoadTranslationsFromMap("en", "English", translations)
	tt.NoError(err)

	tt.Equal("Hello World", T("global.hello"))
	tt.Equal("Goodbye", T("global.bye"))

	err = AddCustomTranslation("en", "global.welcome", "Welcome!")
	tt.NoError(err)

	tt.Equal("Welcome!", T("global.welcome"))

	err = LoadTranslationsFromMap("fr", "French", map[string]string{"bonjour": "Bonjour"})
	tt.NoError(err)

	tt.Equal("Bonjour", TWithLang("fr", "bonjour"))
}
