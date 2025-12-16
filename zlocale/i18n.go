// Package zlocale provides internationalization (i18n) functionality for Go applications.
// It supports multiple languages, named translation tables, parameterized translations,
// and fallback mechanisms. The module is designed to be lightweight and performant,
// utilizing buffer pools from zutil for efficient string operations.
package zlocale

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/zutil"
)

var (
	// ErrLanguageNotFound is returned when a requested language is not available
	ErrLanguageNotFound = fmt.Errorf("language not found")

	// ErrKeyNotFound is returned when a translation key is not found
	ErrKeyNotFound = fmt.Errorf("translation key not found")

	// Global i18n instance for convenience functions
	defaultI18n *I18n
	defaultOnce sync.Once
)

// languageNames maps language codes to their display names
var languageNames = map[string]string{
	"en":    "English",
	"zh-CN": "简体中文",
	"ja":    "日本語",
	"zh":    "简体中文",
}

// Language represents a language with its translation data.
// It provides thread-safe access to translations and cached template processing.
type Language struct {
	// templateCache is the cache system used for template processing
	templateCache TemplateCache
	// data contains the translation key-value pairs
	data map[string]string
	// dataMutex protects concurrent access to data map using standard RWMutex to satisfy the race detector
	dataMutex sync.RWMutex
	// mutex protects concurrent access to the translation data using read-biased mutex for performance
	mutex *zsync.RBMutex

	// Legacy template cache for backward compatibility
	templates map[string]*zstring.Template
	// templateMutex protects concurrent access to the legacy templates cache
	templateMutex *zsync.RBMutex
	// code is the language code (e.g., "en", "zh-CN", "ja")
	code string
	// name is the human-readable language name (e.g., "English", "简体中文")
	name string
	// templateCount tracks the number of cached templates for LRU management
	templateCount int
	// maxTemplates limits the cache size to prevent memory leaks
	maxTemplates int
	// Flag to determine which cache system to use
	useFastCache bool
}

// I18n manages multiple languages and provides translation functionality
type I18n struct {
	// languages contains all loaded languages
	languages map[string]*Language
	// mutex protects concurrent access to the languages map using read-biased mutex for performance
	mutex *zsync.RBMutex
	// defaultLang is the fallback language code
	defaultLang string
	// currentLang is the currently active language code
	currentLang string
}

// New creates a new I18n instance with the specified default language
func New(defaultLang string) *I18n {
	i18n := &I18n{
		defaultLang: defaultLang,
		currentLang: defaultLang,
		languages:   make(map[string]*Language),
		mutex:       zsync.NewRBMutex(),
	}

	return i18n
}

// NewDefault creates a new I18n instance with "en" as the default language
func NewDefault() *I18n {
	return New("en")
}

// getDefault returns the global default i18n instance
func getDefault() *I18n {
	defaultOnce.Do(func() {
		defaultI18n = NewDefault()
	})
	return defaultI18n
}

// LoadLanguage loads a language with translation data
func (i *I18n) LoadLanguage(langCode, langName string, data map[string]string) error {
	return i.LoadLanguageWithConfig(langCode, langName, data, nil)
}

// LoadLanguageWithConfig loads a language with translation data and cache configuration
func (i *I18n) LoadLanguageWithConfig(langCode, langName string, data map[string]string, cache TemplateCache) error {
	if langCode == "" || data == nil {
		return fmt.Errorf("language code and data cannot be empty")
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	if existingLang, exists := i.languages[langCode]; exists {
		existingLang.mutex.Lock()
		defer existingLang.mutex.Unlock()

		if langName != "" {
			existingLang.name = langName
		}

		if cache != nil {
			existingLang.templateCache = cache
		}

		existingLang.dataMutex.Lock()
		for k, v := range data {
			existingLang.data[k] = v
		}
		existingLang.dataMutex.Unlock()

		return nil
	}

	lang := &Language{
		code:          langCode,
		name:          langName,
		data:          make(map[string]string, len(data)),
		mutex:         zsync.NewRBMutex(),
		templates:     make(map[string]*zstring.Template),
		templateMutex: zsync.NewRBMutex(),
		templateCount: 0,
		maxTemplates:  1000,
		useFastCache:  true,
	}

	lang.templateCache = cache

	lang.dataMutex.Lock()
	for k, v := range data {
		lang.data[k] = v
	}
	lang.dataMutex.Unlock()

	i.languages[langCode] = lang
	return nil
}

// SetLanguage sets the current active language
func (i *I18n) SetLanguage(langCode string) error {
	token := i.mutex.RLock()
	_, exists := i.languages[langCode]
	i.mutex.RUnlock(token)

	if !exists {
		return fmt.Errorf("%w: %s", ErrLanguageNotFound, langCode)
	}

	i.mutex.Lock()
	i.currentLang = langCode
	i.mutex.Unlock()

	return nil
}

// GetLanguage returns the current active language code
func (i *I18n) GetLanguage() string {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)
	return i.currentLang
}

// GetLoadedLanguages returns a map of loaded language codes and their names
func (i *I18n) GetLoadedLanguages() map[string]string {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)

	result := make(map[string]string, len(i.languages))
	for code, lang := range i.languages {
		result[code] = lang.name
	}
	return result
}

// T translates a key using the current language
func (i *I18n) T(key string, args ...interface{}) string {
	return i.TWithLang(i.GetLanguage(), key, args...)
}

// TWithLang translates a key using the specified language
func (i *I18n) TWithLang(langCode, key string, args ...interface{}) string {
	if key == "" {
		return ""
	}

	translation, lang := i.getTranslationWithLanguage(langCode, key)
	if translation != "" {
		return i.formatTranslationWithTemplate(lang, translation, args...)
	}

	if langCode != i.defaultLang {
		if translation, lang = i.getTranslationWithLanguage(i.defaultLang, key); translation != "" {
			return i.formatTranslationWithTemplate(lang, translation, args...)
		}
	}

	return key
}

// getTranslationWithLanguage retrieves the translation for a specific language and key, also returns the Language object
func (i *I18n) getTranslationWithLanguage(langCode, key string) (string, *Language) {
	token := i.mutex.RLock()
	lang, exists := i.languages[langCode]
	i.mutex.RUnlock(token)

	if !exists {
		return "", nil
	}

	langToken := lang.mutex.RLock()
	defer lang.mutex.RUnlock(langToken)
	lang.dataMutex.RLock()
	defer lang.dataMutex.RUnlock()
	return lang.data[key], lang
}

// getOrCreateTemplate retrieves or creates a cached template for the given translation string
// Uses the new cache system for improved performance and memory management
func (l *Language) getOrCreateTemplate(templateStr string) (*zstring.Template, error) {
	if l.useFastCache && l.templateCache != nil {
		return l.getOrCreateTemplateFastCache(templateStr)
	}
	return l.getOrCreateTemplateLegacy(templateStr)
}

// getOrCreateTemplateFastCache uses the new FastCache system for template caching
func (l *Language) getOrCreateTemplateFastCache(templateStr string) (*zstring.Template, error) {
	if entry, found := l.templateCache.Get(templateStr); found {
		if tmpl, ok := entry.Template.(*zstring.Template); ok {
			return tmpl, nil
		}
	}

	tmpl, err := zstring.NewTemplate(templateStr, "{", "}")
	if err != nil {
		return nil, err
	}
	entry := &TemplateCacheEntry{
		Template: tmpl,
		Created:  ztime.UnixMicro(ztime.Clock()),
		Accessed: ztime.UnixMicro(ztime.Clock()),
		Hits:     0,
	}

	l.templateCache.Set(templateStr, entry)
	return tmpl, nil
}

// getOrCreateTemplateLegacy uses the original map-based caching for backward compatibility
func (l *Language) getOrCreateTemplateLegacy(templateStr string) (*zstring.Template, error) {
	token := l.templateMutex.RLock()
	tmpl, exists := l.templates[templateStr]
	l.templateMutex.RUnlock(token)

	if exists {
		return tmpl, nil
	}

	l.templateMutex.Lock()
	defer l.templateMutex.Unlock()

	if cachedTmpl, exists := l.templates[templateStr]; exists {
		return cachedTmpl, nil
	}

	tmpl, err := zstring.NewTemplate(templateStr, "{", "}")
	if err != nil {
		return nil, err
	}

	if l.templateCount >= l.maxTemplates {
		var keyToRemove string
		for k := range l.templates {
			keyToRemove = k
			break
		}
		if keyToRemove != "" {
			delete(l.templates, keyToRemove)
			l.templateCount--
		}
	}

	l.templates[templateStr] = tmpl
	l.templateCount++
	return tmpl, nil
}

// formatTranslationWithTemplate formats a translation string using optimized template processing
func (i *I18n) formatTranslationWithTemplate(lang *Language, templateStr string, args ...interface{}) string {
	if len(args) == 0 {
		return templateStr
	}

	hasPrintf := strings.Contains(templateStr, "%")
	hasPlaceholders := strings.Contains(templateStr, "{")

	if hasPrintf && !hasPlaceholders {
		buf := zutil.GetBuff()
		defer zutil.PutBuff(buf)
		buf.Reset()
		fmt.Fprintf(buf, templateStr, args...)
		return buf.String()
	}

	if hasPlaceholders {
		tmpl, err := lang.getOrCreateTemplate(templateStr)
		if err == nil {
			buf := zutil.GetBuff()
			defer zutil.PutBuff(buf)
			buf.Reset()

			_, err = tmpl.Process(buf, func(w io.Writer, tag string) (int, error) {
				if tag == "" {
					return 0, nil
				}

				if len(tag) > 0 && tag[0] >= '0' && tag[0] <= '9' {
					index := 0
					validNumber := true
					for _, c := range tag {
						if c >= '0' && c <= '9' {
							index = index*10 + int(c-'0')
						} else {
							validNumber = false
							break
						}
					}

					if validNumber && index < len(args) {
						return w.Write([]byte(fmt.Sprint(args[index])))
					}
				}

				return w.Write([]byte(fmt.Sprintf("{%s}", tag)))
			})

			if err == nil {
				return buf.String()
			}
		}
	}

	buf := zutil.GetBuff()
	defer zutil.PutBuff(buf)
	buf.Reset()

	result := templateStr
	hasReplacements := false
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprint(arg))
			hasReplacements = true
		}
	}

	if !hasReplacements {
		fmt.Fprintf(buf, templateStr, args...)
		return buf.String()
	}

	buf.WriteString(result)
	return buf.String()
}

// HasLanguage checks if a language is loaded
func (i *I18n) HasLanguage(langCode string) bool {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)
	_, exists := i.languages[langCode]
	return exists
}

// HasKey checks if a translation key exists for a specific language
func (i *I18n) HasKey(langCode, key string) bool {
	token := i.mutex.RLock()
	lang, exists := i.languages[langCode]
	i.mutex.RUnlock(token)

	if !exists {
		return false
	}

	langToken := lang.mutex.RLock()
	defer lang.mutex.RUnlock(langToken)
	lang.dataMutex.RLock()
	_, exists = lang.data[key]
	lang.dataMutex.RUnlock()
	return exists
}

// RemoveLanguage removes a language from the i18n instance
func (i *I18n) RemoveLanguage(langCode string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if langCode == i.defaultLang {
		return fmt.Errorf("cannot remove default language")
	}

	if _, exists := i.languages[langCode]; !exists {
		return fmt.Errorf("%w: %s", ErrLanguageNotFound, langCode)
	}

	delete(i.languages, langCode)

	if i.currentLang == langCode {
		i.currentLang = i.defaultLang
	}

	return nil
}

// GetMemoryUsage returns memory usage statistics for monitoring
func (i *I18n) GetMemoryUsage() map[string]interface{} {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)

	stats := make(map[string]interface{})
	totalTranslations := 0
	totalTemplates := 0

	for code, lang := range i.languages {
		langToken := lang.mutex.RLock()
		lang.dataMutex.RLock()
		translationCount := len(lang.data)
		lang.dataMutex.RUnlock()
		lang.mutex.RUnlock(langToken)

		langStats := map[string]interface{}{
			"translations": translationCount,
			"cache_type":   "legacy",
		}

		if lang.templateCache != nil {
			cacheStats := lang.templateCache.Stats()
			langStats["templates"] = map[string]interface{}{
				"total":           cacheStats.TotalItems,
				"hit_count":       cacheStats.HitCount,
				"miss_count":      cacheStats.MissCount,
				"hit_rate":        cacheStats.HitRate,
				"memory_usage":    cacheStats.MemoryUsage,
				"eviction_count":  cacheStats.EvictionCount,
				"cleaner_level":   cacheStats.CleanerLevel,
				"idle_duration":   cacheStats.IdleDuration,
				"cleaner_running": cacheStats.IsCleanerRunning,
			}
			totalTemplates += cacheStats.TotalItems
			langStats["cache_type"] = "fastcache"
		} else {
			templateToken := lang.templateMutex.RLock()
			templateCount := len(lang.templates)
			lang.templateMutex.RUnlock(templateToken)

			langStats["templates"] = map[string]interface{}{
				"total": templateCount,
			}
			totalTemplates += templateCount
		}

		stats[code] = langStats
	}

	stats["total"] = map[string]interface{}{
		"languages":    len(i.languages),
		"translations": totalTranslations,
		"templates":    totalTemplates,
	}

	return stats
}

// GetCacheStats returns detailed cache statistics for all languages
func (i *I18n) GetCacheStats() map[string]CacheStats {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)

	stats := make(map[string]CacheStats)

	for code, lang := range i.languages {
		if lang.templateCache != nil {
			stats[code] = lang.templateCache.Stats()
		}
	}

	return stats
}

// ClearTemplateCache clears the template cache for all languages to free memory
func (i *I18n) ClearTemplateCache() {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)

	for _, lang := range i.languages {
		if lang.templateCache != nil {
			lang.templateCache.Clear()
		} else {
			lang.templateMutex.Lock()
			lang.templates = make(map[string]*zstring.Template)
			lang.templateCount = 0
			lang.templateMutex.Unlock()
		}
	}
}

// Close closes all cache resources and stops background cleaners
func (i *I18n) Close() {
	token := i.mutex.RLock()
	defer i.mutex.RUnlock(token)

	for _, lang := range i.languages {
		if lang.templateCache != nil {
			lang.templateCache.Close()
		}
	}
}

// LoadLanguage loads a language into the global i18n instance
func LoadLanguage(langCode, langName string, data map[string]string) error {
	return getDefault().LoadLanguage(langCode, langName, data)
}

// SetLanguage sets the active language for the global i18n instance
func SetLanguage(langCode string) error {
	return getDefault().SetLanguage(langCode)
}

// GetLanguage returns the current active language from the global i18n instance
func GetLanguage() string {
	return getDefault().GetLanguage()
}

// T translates a key using the global i18n instance
func T(key string, args ...interface{}) string {
	return getDefault().T(key, args...)
}

// TWithLang translates a key using the specified language in the global i18n instance
func TWithLang(langCode, key string, args ...interface{}) string {
	return getDefault().TWithLang(langCode, key, args...)
}

// HasLanguage checks if a language is loaded in the global i18n instance
func HasLanguage(langCode string) bool {
	return getDefault().HasLanguage(langCode)
}

// HasKey checks if a translation key exists in the global i18n instance
func HasKey(langCode, key string) bool {
	return getDefault().HasKey(langCode, key)
}

// GetLoadedLanguages returns all loaded languages from the global i18n instance
func GetLoadedLanguages() map[string]string {
	return getDefault().GetLoadedLanguages()
}
