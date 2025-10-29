package zlocale

import (
	"strings"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestBasicFunctionality(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i18n := New("en")

	data := map[string]string{
		"hello":     "Hello",
		"goodbye":   "Goodbye",
		"user.name": "John Doe",
	}

	err := i18n.LoadLanguage("en", "English", data)
	tt.NoError(err)

	tt.Log(i18n.GetMemoryUsage())
	tt.Log(i18n.GetLoadedLanguages())

	tt.EqualTrue(i18n.HasLanguage("en"))
	tt.EqualFalse(i18n.HasLanguage("zh-CN"))

	tt.Equal(i18n.T("hello"), "Hello")

	templateData := map[string]string{
		"user.welcome": "Welcome, {0}!",
	}
	i18n.LoadLanguage("en", "English", templateData)

	tt.Equal(i18n.T("user.welcome", "Alice"), "Welcome, Alice!")
}

func TestConcurrentTemplateAccess(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i18n := New("en")

	data := map[string]string{
		"template1": "Hello {0}!",
		"template2": "Welcome {0}, you have {1} messages",
		"template3": "Simple text without args",
	}

	i18n.LoadLanguage("en", "English", data)

	var wg sync.WaitGroup
	concurrency := 100
	iterations := 1000

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tt.Equal(i18n.T("template1", "User"), "Hello User!")
				tt.Equal(i18n.T("template2", "Alice", 5), "Welcome Alice, you have 5 messages")
				tt.Equal(i18n.T("template3"), "Simple text without args")
			}
		}(i)
	}

	wg.Wait()

	tt.Equal(i18n.T("template1", "Test"), "Hello Test!")
}

func TestMemoryUsageManagement(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i18n := New("en")

	data := map[string]string{
		"test": "Value {0}",
	}
	i18n.LoadLanguage("en", "English", data)
	initialStats := i18n.GetMemoryUsage()
	tt.NotNil(initialStats)

	for i := 0; i < 1500; i++ {
		templateKey := "dynamic" + string(rune(i))
		i18n.T(templateKey, i)
	}
	afterStats := i18n.GetMemoryUsage()
	tt.NotNil(afterStats)

	i18n.ClearTemplateCache()
	clearedStats := i18n.GetMemoryUsage()
	tt.NotNil(clearedStats)

	tt.Logf("Initial stats: %+v", initialStats)
	tt.Logf("After template creation: %+v", afterStats)
	tt.Logf("After cache clear: %+v", clearedStats)
}

func TestRaceConditionFix(t *testing.T) {
	tt := zlsgo.NewTest(t)

	i18n := New("en")
	data := map[string]string{
		"race_test": "Hello {0}!",
	}
	i18n.LoadLanguage("en", "English", data)

	var wg sync.WaitGroup
	concurrency := 50
	iterations := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tt.Equal(i18n.T("race_test", "User"), "Hello User!")
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkTemplateCaching(b *testing.B) {
	i18n := New("en")
	data := map[string]string{
		"benchmark": "Hello {0}, welcome to {1}!",
	}
	i18n.LoadLanguage("en", "English", data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i18n.T("benchmark", "User", "Go")
		}
	})
}

func BenchmarkConcurrentAccess(b *testing.B) {
	i18n := New("en")
	data := map[string]string{
		"simple": "Hello World",
		"param":  "Hello {0}!",
	}
	i18n.LoadLanguage("en", "English", data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i18n.T("simple")
			i18n.T("param", "Test")
		}
	})
}

func TestNewDefault(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := NewDefault()
	tt.NotNil(i18n)
	tt.Equal("en", i18n.GetLanguage())
	languages := i18n.GetLoadedLanguages()
	tt.Equal(0, len(languages))
}

func TestSetLanguage(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	i18n.LoadLanguage("en", "English", map[string]string{"hello": "Hello"})
	i18n.LoadLanguage("fr", "French", map[string]string{"hello": "Bonjour"})
	
	err := i18n.SetLanguage("fr")
	tt.NoError(err)
	tt.Equal("fr", i18n.GetLanguage())
	tt.Equal("Bonjour", i18n.T("hello"))
	
	err = i18n.SetLanguage("de")
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "language not found"))
	
	tt.Equal("fr", i18n.GetLanguage())
}

func TestGetLanguage(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("zh-CN")
	tt.Equal("zh-CN", i18n.GetLanguage())
	
	i18n.LoadLanguage("zh-CN", "Chinese", map[string]string{"test": "测试"})
	tt.Equal("zh-CN", i18n.GetLanguage())
	
	i18n.LoadLanguage("en", "English", map[string]string{"test": "Test"})
	i18n.SetLanguage("en")
	tt.Equal("en", i18n.GetLanguage())
}

func TestTWithLang(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	i18n.LoadLanguage("en", "English", map[string]string{
		"hello": "Hello",
		"welcome": "Welcome {0}",
	})
	
	i18n.LoadLanguage("fr", "French", map[string]string{
		"hello": "Bonjour",
		"welcome": "Bienvenue {0}",
	})
	
	tt.Equal("Hello", i18n.TWithLang("en", "hello"))
	tt.Equal("Bonjour", i18n.TWithLang("fr", "hello"))
	tt.Equal("Welcome Alice", i18n.TWithLang("en", "welcome", "Alice"))
	tt.Equal("Bienvenue Alice", i18n.TWithLang("fr", "welcome", "Alice"))
	
	tt.Equal("Hello", i18n.TWithLang("de", "hello"))
	
	tt.Equal("", i18n.TWithLang("en", ""))
}

func TestHasKey(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	tt.EqualFalse(i18n.HasKey("en", "hello"))
	
	i18n.LoadLanguage("en", "English", map[string]string{
		"hello": "Hello",
		"user.name": "John",
	})
	
	tt.EqualTrue(i18n.HasKey("en", "hello"))
	tt.EqualTrue(i18n.HasKey("en", "user.name"))
	tt.EqualFalse(i18n.HasKey("en", "nonexistent"))
	
	tt.EqualFalse(i18n.HasKey("fr", "hello"))
}

func TestRemoveLanguage(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	i18n.LoadLanguage("en", "English", map[string]string{"hello": "Hello"})
	i18n.LoadLanguage("fr", "French", map[string]string{"hello": "Bonjour"})
	i18n.LoadLanguage("de", "German", map[string]string{"hello": "Hallo"})
	
	i18n.SetLanguage("fr")
	tt.Equal("fr", i18n.GetLanguage())
	
	err := i18n.RemoveLanguage("de")
	tt.NoError(err)
	tt.EqualFalse(i18n.HasLanguage("de"))
	
	tt.Equal("fr", i18n.GetLanguage())
	
	err = i18n.RemoveLanguage("fr")
	tt.NoError(err)
	tt.EqualFalse(i18n.HasLanguage("fr"))
	tt.Equal("en", i18n.GetLanguage())
	
	err = i18n.RemoveLanguage("en")
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "cannot remove default language"))
	
	err = i18n.RemoveLanguage("ja")
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "language not found"))
}

func TestGetCacheStats(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	stats := i18n.GetCacheStats()
	tt.NotNil(stats)
	tt.Equal(0, len(stats))
	
	i18n.LoadLanguage("en", "English", map[string]string{"test": "Test {0}"})
	stats = i18n.GetCacheStats()
	tt.Equal(0, len(stats))
	
	cache := NewLegacyCacheAdapter(100)
	i18n.LoadLanguageWithConfig("fr", "French", 
		map[string]string{"test": "Test {0}"}, cache)
	
	stats = i18n.GetCacheStats()
	tt.Equal(1, len(stats))
	tt.NotNil(stats["fr"])
	
	i18n.TWithLang("fr", "test", "World")
	frStats := stats["fr"]
	tt.NotNil(frStats)
}

func TestClose(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	i18n.LoadLanguage("en", "English", map[string]string{"test": "Test"})
	
	cache := NewLegacyCacheAdapter(100)
	i18n.LoadLanguageWithConfig("fr", "French", 
		map[string]string{"test": "Test"}, cache)
	
	i18n.TWithLang("fr", "test")
	
	i18n.Close()
	
	tt.Equal("Test", i18n.T("test"))
}

func TestLoadLanguageWithConfig(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	cache := NewLegacyCacheAdapter(50)
	data := map[string]string{"hello": "Hello", "welcome": "Welcome {0}"}
	
	err := i18n.LoadLanguageWithConfig("en", "English", data, cache)
	tt.NoError(err)
	tt.EqualTrue(i18n.HasLanguage("en"))
	
	result := i18n.T("welcome", "World")
	tt.Equal("Welcome World", result)
	
	err = i18n.LoadLanguageWithConfig("fr", "French", 
		map[string]string{"bonjour": "Bonjour"}, nil)
	tt.NoError(err)
	tt.Equal("Bonjour", i18n.TWithLang("fr", "bonjour"))
	
	err = i18n.LoadLanguageWithConfig("", "English", data, nil)
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "cannot be empty"))
	
	err = i18n.LoadLanguageWithConfig("de", "German", nil, nil)
	tt.EqualTrue(err != nil)
	tt.EqualTrue(strings.Contains(err.Error(), "cannot be empty"))
}

func TestFastCachePath(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	i18n := New("en")
	
	cache := NewLegacyCacheAdapter(100)
	data := map[string]string{
		"template": "Hello {0}!",
		"simple": "Simple text",
	}
	
	err := i18n.LoadLanguageWithConfig("en", "English", data, cache)
	tt.NoError(err)
	
	result := i18n.T("template", "World")
	tt.Equal("Hello World!", result)
	
	result = i18n.T("simple")
	tt.Equal("Simple text", result)
}

func TestGlobalFunctions(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	defaultI18n = nil
	defaultOnce = sync.Once{}
	
	err := LoadLanguage("en", "English", map[string]string{
		"global.hello": "Hello World",
		"global.welcome": "Welcome {0}",
	})
	tt.NoError(err)
	
	tt.Equal("Hello World", T("global.hello"))
	tt.Equal("Welcome Alice", T("global.welcome", "Alice"))
	
	err = SetLanguage("en")
	tt.NoError(err)
	tt.Equal("en", GetLanguage())
	
	tt.EqualTrue(HasLanguage("en"))
	tt.EqualTrue(HasKey("en", "global.hello"))
	tt.EqualFalse(HasKey("en", "nonexistent"))
	
	languages := GetLoadedLanguages()
	tt.NotNil(languages)
	tt.Equal("English", languages["en"])
	
	tt.Equal("Hello World", TWithLang("en", "global.hello"))
	
	LoadLanguage("fr", "French", map[string]string{"fallback": "Fallback"})
	tt.Equal("Hello World", TWithLang("de", "global.hello"))
}