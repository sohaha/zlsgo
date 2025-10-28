package zlocale

import (
	"fmt"
)

// Loader provides functionality to load translations from maps
// Simplified version that only supports embedded translations
type Loader struct {
	i18n *I18n
}

// NewLoader creates a new loader for the given i18n instance
func NewLoader(i18n *I18n) *Loader {
	return &Loader{i18n: i18n}
}

// LoadTranslationsFromMap loads translations from a map[string]string
func (l *Loader) LoadTranslationsFromMap(langCode, langName string, translations map[string]string) error {
	if translations == nil {
		return fmt.Errorf("translations map cannot be nil")
	}

	return l.i18n.LoadLanguage(langCode, langName, translations)
}

// AddCustomTranslation adds a custom translation to an existing language
// This allows runtime extension of the built-in translations
func (l *Loader) AddCustomTranslation(langCode, key, value string) error {
	langName := langCode
	if name, ok := languageNames[langCode]; ok {
		langName = name
	}

	customTranslations := map[string]string{key: value}
	return l.i18n.LoadLanguage(langCode, langName, customTranslations)
}

// GetAvailableLanguages returns the list of available language codes
func (l *Loader) GetAvailableLanguages() []string {
	languages := l.i18n.GetLoadedLanguages()
	result := make([]string, 0, len(languages))
	for code := range languages {
		result = append(result, code)
	}
	return result
}

// LoadTranslationsFromMap loads translations from a map using the global i18n instance
func LoadTranslationsFromMap(langCode, langName string, translations map[string]string) error {
	return NewLoader(getDefault()).LoadTranslationsFromMap(langCode, langName, translations)
}

// AddCustomTranslation adds a custom translation using the global i18n instance
func AddCustomTranslation(langCode, key, value string) error {
	return NewLoader(getDefault()).AddCustomTranslation(langCode, key, value)
}
