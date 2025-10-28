package zlocale

import (
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
