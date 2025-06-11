package zlog

import (
	"io"
	"testing"
)

func BenchmarkLevelLogging(b *testing.B) {
	logger := NewZLog(io.Discard, "", BitDefault, LogDump, false, 3)

	b.Run("Debug", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Debug("This is a debug message", i)
		}
	})

	b.Run("Info", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("This is an info message", i)
		}
	})

	b.Run("Error", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Error("This is an error message", i)
		}
	})
}

func BenchmarkFormattedLogging(b *testing.B) {
	logger := NewZLog(io.Discard, "", BitDefault, LogDump, false, 3)

	b.Run("Debugf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Debugf("This is a debug message: %d", i)
		}
	})

	b.Run("Infof", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Infof("This is an info message: %d", i)
		}
	})

	b.Run("Errorf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Errorf("This is an error message: %d", i)
		}
	})
}

func BenchmarkComplexLogging(b *testing.B) {
	logger := NewZLog(io.Discard, "", BitDefault, LogDump, false, 3)

	type complexStruct struct {
		Name   string
		Age    int
		Scores []float64
		Tags   map[string]string
	}

	data := complexStruct{
		Name:   "Test User",
		Age:    30,
		Scores: []float64{95.5, 87.3, 91.0},
		Tags: map[string]string{
			"role":     "admin",
			"location": "server1",
		},
	}

	b.Run("Dump", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Dump(data)
		}
	})

	b.Run("MultipleArgs", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("User data:", data, "Index:", i)
		}
	})
}

func BenchmarkHeaderFlags(b *testing.B) {
	b.Run("StdFlag", func(b *testing.B) {
		logger := NewZLog(io.Discard, "", BitStdFlag, LogDump, false, 3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("This is a test message")
		}
	})

	b.Run("DefaultFlag", func(b *testing.B) {
		logger := NewZLog(io.Discard, "", BitDefault, LogDump, false, 3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("This is a test message")
		}
	})

	b.Run("AllFlags", func(b *testing.B) {
		logger := NewZLog(io.Discard, "", BitDate|BitTime|BitMicroSeconds|BitLongFile|BitLevel, LogDump, false, 3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("This is a test message")
		}
	})
}
