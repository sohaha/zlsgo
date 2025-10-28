package zlocale

import (
	"fmt"
	"testing"
)

func BenchmarkCache(b *testing.B) {
	b.Run("Legacy", func(b *testing.B) {
		i18n := New("en")

		data := make(map[string]string)
		for i := 0; i < 1000; i++ {
			data[fmt.Sprintf("template_%d", i)] = fmt.Sprintf("Hello {0}, this is template %d", i)
		}

		err := i18n.LoadLanguageWithConfig("en", "English", data, NewLegacyCacheAdapter(1000))
		if err != nil {
			b.Fatalf("Failed to load language: %v", err)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("template_%d", i%200)
					i18n.T(key, "World")
				}
			}
		})
	})
}
