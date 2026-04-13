package zcli

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	zls "github.com/sohaha/zlsgo"
)

func TestProgressBarBasic(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(100,
		func(o *ProgressOptions) {
			o.Writer = &buf
			o.Width = 8
			o.FlushInterval = 0
		},
	)
	pb.Set(25)
	pb.Add(25)
	pb.Add(50)
	pb.Finish()

	out := buf.String()
	tt.EqualTrue(strings.Contains(out, "100%"))
	tt.EqualTrue(strings.Contains(out, "Elapsed"))
	tt.EqualTrue(strings.HasSuffix(out, "\n"))
	tt.EqualTrue(!strings.Contains(out, "\r"))
}

func TestProgressBarConcurrent(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(100,
		func(o *ProgressOptions) {
			o.Writer = &buf
			o.FlushInterval = 0
		},
	)

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				pb.Add(1)
			}
		}()
	}
	wg.Wait()
	pb.Finish()

	tt.EqualTrue(strings.Contains(buf.String(), "100%"))
	tt.Equal(int64(100), pb.Current())
}

func TestProgressBarSpinner(t *testing.T) {
	tt := zls.NewTest(t)
	var writer captureWriter
	pb := NewProgressBar(0,
		func(o *ProgressOptions) {
			o.Writer = &writer
			o.FlushInterval = 0
			o.Spinner = []rune{'-', '\\', '|'}
		},
	)
	pb.Add(1)
	pb.Add(1)
	pb.Add(1)
	pb.Finish()

	seen := map[string]bool{}
	for _, w := range writer.writes {
		switch {
		case strings.Contains(w, "[-]"):
			seen["-"] = true
		case strings.Contains(w, "[\\]"):
			seen["\\"] = true
		case strings.Contains(w, "[|]"):
			seen["|"] = true
		}
	}
	tt.EqualTrue(seen["-"])
	tt.EqualTrue(seen["|"])
}

func TestProgressBarClampAndString(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(10,
		func(o *ProgressOptions) {
			o.Writer = &buf
			o.Prefix = "sync"
			o.Suffix = "done"
		},
	)
	pb.Add(100)
	tt.Equal(int64(10), pb.Current())
	tt.Equal(int64(10), pb.Total())
	tt.EqualTrue(strings.Contains(pb.String(), "100% 10/10"))
	tt.EqualTrue(strings.Contains(pb.String(), "sync"))
	tt.EqualTrue(strings.Contains(pb.String(), "done"))
	_ = pb.Close()
}

func TestProgressBarFlushIntervalZeroWritesEveryChange(t *testing.T) {
	tt := zls.NewTest(t)
	var writer captureWriter
	pb := NewProgressBar(100,
		func(o *ProgressOptions) {
			o.Writer = &writer
			o.FlushInterval = 0
		},
	)
	pb.Add(1)
	pb.Add(1)
	pb.Finish()
	tt.Equal(2, len(writer.writes))
	tt.EqualTrue(strings.Contains(writer.writes[0], "1/100"))
	tt.EqualTrue(strings.Contains(writer.writes[1], "2/100"))
}

func TestProgressBarFinishDoesNotDuplicateFinalNonTerminalLine(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(1,
		func(o *ProgressOptions) {
			o.Writer = &buf
			o.FlushInterval = 0
		},
	)
	pb.Add(1)
	pb.Finish()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	tt.Equal(1, len(lines))
	tt.EqualTrue(strings.Contains(lines[0], "100% 1/1"))
}

func TestProgressBarFinishRefreshesFinalElapsed(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(1, func(o *ProgressOptions) {
		o.Writer = &buf
		o.FlushInterval = 0
	})
	pb.Add(1)
	pb.start = time.Now().Add(-2 * time.Second)
	pb.Finish()

	out := strings.TrimSpace(buf.String())
	tt.EqualTrue(strings.Contains(out, "Elapsed 2s") || strings.Contains(out, "Elapsed 1s"))
}

func TestProgressBarStringDoesNotAdvanceSpinner(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(0,
		func(o *ProgressOptions) {
			o.Writer = &buf
			o.Spinner = []rune{'-', '\\', '|'}
		},
	)
	pb.Set(1)
	first := pb.String()
	second := pb.String()
	tt.Equal(first, second)
	_ = pb.Close()
}

func TestProgressBarIncrementAndSetTotal(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(5, func(o *ProgressOptions) {
		o.Writer = &buf
		o.FlushInterval = 0
	})
	pb.Increment()
	pb.Increment()
	tt.Equal(int64(2), pb.Current())
	pb.SetTotal(1)
	tt.Equal(int64(1), pb.Current())
	tt.Equal(int64(1), pb.Total())
	pb.SetTotal(0)
	tt.Equal(int64(0), pb.Total())
}

func TestProgressBarSetTotalRerendersWhenCurrentUnchanged(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(10, func(o *ProgressOptions) {
		o.Writer = &buf
		o.FlushInterval = 0
	})
	pb.Set(5)

	buf.Reset()
	pb.SetTotal(20)

	out := buf.String()
	tt.EqualTrue(strings.Contains(out, "25% 5/20"))
}

func TestProgressBarSetTotalRerendersWithinFlushInterval(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(10, func(o *ProgressOptions) {
		o.Writer = &buf
		o.FlushInterval = time.Hour
	})
	pb.Set(5)

	buf.Reset()
	pb.SetTotal(20)

	out := buf.String()
	tt.EqualTrue(strings.Contains(out, "25% 5/20"))
}

func TestProgressBarZeroFillAndEmptyFallbackToDefaults(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(2, func(o *ProgressOptions) {
		o.Writer = &buf
		o.Width = 4
		o.Fill = 0
		o.Empty = 0
		o.FlushInterval = 0
	})
	pb.Add(1)
	pb.Finish()

	out := buf.String()
	tt.EqualTrue(strings.Contains(out, "[==> ]") || strings.Contains(out, "[====]"))
	tt.EqualTrue(!strings.ContainsRune(out, rune(0)))
}

func TestProgressBarSmoothedETA(t *testing.T) {
	tt := zls.NewTest(t)
	pb := NewProgressBar(100)
	pb.start = time.Now().Add(-10 * time.Second)
	pb.lastSampleTime = pb.start
	pb.lastSampleCurrent = 0
	meta := pb.knownProgressMeta(50, 50, pb.start.Add(5*time.Second), true)
	tt.EqualTrue(strings.Contains(meta, "ETA"))
	tt.EqualTrue(pb.smoothedPerSecond > 0)
}

func TestProgressBarEstimatedETAFallback(t *testing.T) {
	tt := zls.NewTest(t)
	pb := NewProgressBar(100)
	eta, ok := pb.estimatedETA(50, 100, 10*time.Second)
	tt.EqualTrue(ok)
	tt.Equal(10*time.Second, eta)
}

type captureWriter struct {
	mu     sync.Mutex
	writes []string
}

func (c *captureWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	c.writes = append(c.writes, string(p))
	c.mu.Unlock()
	return len(p), nil
}

func TestFitProgressBarWidth(t *testing.T) {
	tt := zls.NewTest(t)
	tt.Equal(40, fitProgressBarWidth(40, 20, 80))
	tt.Equal(16, fitProgressBarWidth(40, 20, 39))
	tt.Equal(0, fitProgressBarWidth(40, 20, 25))
}

func TestEnvTerminalWidth(t *testing.T) {
	tt := zls.NewTest(t)
	old := os.Getenv("COLUMNS")
	defer func() {
		_ = os.Setenv("COLUMNS", old)
	}()
	_ = os.Setenv("COLUMNS", "72")
	width, ok := envTerminalWidth()
	tt.EqualTrue(ok)
	tt.Equal(72, width)
}

func TestStringDisplayWidth(t *testing.T) {
	tt := zls.NewTest(t)
	tt.Equal(4, stringDisplayWidth("\u4e0a\u4f20"))
	tt.Equal(6, stringDisplayWidth("\u4e0a\u4f20go"))
	tt.Equal(2, stringDisplayWidth("\U0001f600"))
	tt.Equal(2, stringDisplayWidth("👍🏻"))
	tt.Equal(2, stringDisplayWidth("1️⃣"))
	tt.Equal(2, stringDisplayWidth("👨‍👩‍👧‍👦"))
}

func TestFitProgressLinePrefersCoreAndPrefix(t *testing.T) {
	tt := zls.NewTest(t)
	line := fitProgressLine("超长前缀任务", "[====] 50% 5/10", "后缀END", 20)
	tt.EqualTrue(strings.Contains(line, "[====]"))
	tt.EqualTrue(strings.Contains(line, "50%"))
	tt.EqualTrue(strings.HasPrefix(line, "..."))
	tt.EqualTrue(!strings.HasSuffix(line, "后缀END"))
	tt.EqualTrue(stringDisplayWidth(line) <= 20)
}

func TestFitKnownProgressCorePrefersMetaOverBar(t *testing.T) {
	tt := zls.NewTest(t)
	core := fitKnownProgressCore("[==========]", "50% 5/10 ETA 10s Elapsed 12s", 12)
	tt.EqualTrue(strings.Contains(core, "50%"))
	tt.EqualTrue(strings.Contains(core, "5/10"))
	tt.EqualTrue(!strings.HasPrefix(core, "[===="))
	tt.EqualTrue(stringDisplayWidth(core) <= 12)
}

func TestTruncateDisplayWidthWithEmojiCluster(t *testing.T) {
	tt := zls.NewTest(t)
	got := truncateDisplayWidth("👨‍👩‍👧‍👦family", 7)
	tt.EqualTrue(strings.HasPrefix(got, "👨‍👩‍👧‍👦"))
	tt.EqualTrue(strings.HasSuffix(got, "..."))
	tt.EqualTrue(stringDisplayWidth(got) <= 7)
}

func TestShortenProgressMetaKeepsCoreFields(t *testing.T) {
	tt := zls.NewTest(t)
	meta := " 50% 5/10 ETA 10s Elapsed 12s"
	got := shortenProgressMeta(meta, 16)
	tt.EqualTrue(strings.Contains(got, "50%"))
	tt.EqualTrue(strings.Contains(got, "5/10"))
	tt.EqualTrue(!strings.Contains(got, "Elapsed"))
	tt.EqualTrue(stringDisplayWidth(got) <= 16)
}

func TestProgressBarStringWithEmojiPrefixKeepsBarVisible(t *testing.T) {
	tt := zls.NewTest(t)
	var buf bytes.Buffer
	pb := NewProgressBar(10, func(o *ProgressOptions) {
		o.Writer = &buf
		o.Prefix = "👍🏻"
		o.Width = 8
	})
	pb.Set(5)
	out := pb.String()
	tt.EqualTrue(strings.Contains(out, "["))
	tt.EqualTrue(strings.Contains(out, "]"))
}

func TestProgressBarSpinnerLinePrefersCore(t *testing.T) {
	tt := zls.NewTest(t)
	line := fitProgressLine("超长前缀任务", "[|] 12", "尾部说明", 10)
	tt.EqualTrue(strings.Contains(line, "[|]"))
	tt.EqualTrue(strings.Contains(line, "12"))
	tt.EqualTrue(stringDisplayWidth(line) <= 10)
}

func TestProgressBarDemo(t *testing.T) {
	const total = 50
	step := 20 * time.Millisecond

	t.Log("=== 普通进度条 ===")
	pb := NewProgressBar(total, func(o *ProgressOptions) {
		o.Writer = os.Stdout
		o.Prefix = "下载"
		o.Suffix = "MB/s"
		o.Width = 30
	})
	for i := 0; i < total; i++ {
		time.Sleep(step)
		pb.Increment()
	}
	pb.Finish()

	time.Sleep(200 * time.Millisecond)

	t.Log("=== 自定义填充字符 ===")
	pb2 := NewProgressBar(total, func(o *ProgressOptions) {
		o.Writer = os.Stdout
		o.Prefix = "上传"
		o.Fill = '#'
		o.Empty = '.'
		o.Width = 30
	})
	for i := 0; i < total; i++ {
		time.Sleep(step)
		pb2.Increment()
	}
	pb2.Finish()

	time.Sleep(200 * time.Millisecond)

	t.Log("=== Spinner（未知总量）===")
	pb3 := NewProgressBar(0, func(o *ProgressOptions) {
		o.Writer = os.Stdout
		o.Prefix = "处理"
		o.Spinner = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	})
	for i := 0; i < total; i++ {
		time.Sleep(step)
		pb3.Increment()
	}
	pb3.Finish()
}
