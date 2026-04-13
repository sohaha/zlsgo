package zcli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zutil"
)

const (
	defaultProgressWidth     = 40
	defaultFlushInterval     = 120 * time.Millisecond
	defaultSpinnerChars      = "|/-\\"
	defaultProgressFillChar  = '='
	defaultProgressEmptyChar = ' '
	minAdaptiveBarWidth      = 8
	etaSmoothingAlpha        = 0.25
)

var progressOutputMu sync.Mutex

type ProgressOptions struct {
	Writer        io.Writer
	Width         int
	Prefix        string
	Suffix        string
	Fill          byte
	Empty         byte
	Spinner       []rune
	FlushInterval time.Duration
}

type ProgressBar struct {
	writer        io.Writer
	total         *zutil.Int64
	width         int
	prefix        string
	suffix        string
	fill          byte
	empty         byte
	spinChars     []rune
	flushInterval time.Duration

	start             time.Time
	lastRender        time.Time
	lastPercent       int
	lastCurrent       int64
	lastTotal         int64
	lastSampleTime    time.Time
	lastSampleCurrent int64
	smoothedPerSecond float64
	lastLine          string

	current     *zutil.Int64
	done        *zutil.Bool
	interactive bool
	mu          sync.Mutex
}

func NewProgressBar(total int64, opts ...func(*ProgressOptions)) *ProgressBar {
	now := time.Now()
	o := ProgressOptions{
		Writer:        os.Stdout,
		Width:         defaultProgressWidth,
		Fill:          defaultProgressFillChar,
		Empty:         defaultProgressEmptyChar,
		Spinner:       []rune(defaultSpinnerChars),
		FlushInterval: defaultFlushInterval,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	p := &ProgressBar{
		writer:         o.Writer,
		total:          zutil.NewInt64(normalizeTotal(total)),
		width:          o.Width,
		prefix:         o.Prefix,
		suffix:         o.Suffix,
		fill:           o.Fill,
		empty:          o.Empty,
		spinChars:      append([]rune(nil), o.Spinner...),
		flushInterval:  o.FlushInterval,
		start:          now,
		lastSampleTime: now,
		current:        zutil.NewInt64(0),
		done:           zutil.NewBool(false),
	}
	if p.width <= 0 {
		p.width = defaultProgressWidth
	}
	if p.fill == 0 {
		p.fill = defaultProgressFillChar
	}
	if p.empty == 0 {
		p.empty = defaultProgressEmptyChar
	}
	if p.flushInterval < 0 {
		p.flushInterval = 0
	}
	if len(p.spinChars) == 0 {
		p.spinChars = []rune(defaultSpinnerChars)
	}
	if p.writer == nil {
		p.writer = os.Stdout
	}
	p.interactive = isTerminalWriter(p.writer)
	return p
}

func (p *ProgressBar) Add(delta int64) {
	if delta == 0 || p.done.Load() {
		return
	}

	for {
		current := p.current.Load()
		next := p.normalizeCurrent(current + delta)
		if p.current.CAS(current, next) {
			if p.done.Load() {
				return
			}
			p.render(next, false)
			return
		}
	}
}

func (p *ProgressBar) Increment() {
	p.Add(1)
}

func (p *ProgressBar) Set(value int64) {
	if p.done.Load() {
		return
	}
	value = p.normalizeCurrent(value)
	p.current.Store(value)
	if p.done.Load() {
		return
	}
	p.render(value, false)
}

func (p *ProgressBar) SetTotal(total int64) {
	total = normalizeTotal(total)
	p.total.Store(total)
	current := p.normalizeCurrent(p.current.Load())
	p.current.Store(current)
	if p.done.Load() {
		return
	}
	p.render(current, false)
}

func (p *ProgressBar) Current() int64 {
	return p.current.Load()
}

func (p *ProgressBar) Total() int64 {
	return p.total.Load()
}

func (p *ProgressBar) String() string {
	current := p.normalizeCurrent(p.current.Load())
	now := time.Now()
	p.mu.Lock()
	meta := p.knownProgressMeta(current, p.percent(current), now, false)
	p.mu.Unlock()
	return p.format(current, meta)
}

func (p *ProgressBar) Close() error {
	p.Finish()
	return nil
}

func (p *ProgressBar) Finish() {
	if !p.done.CAS(false, true) {
		return
	}
	current := p.normalizeCurrent(p.current.Load())
	p.current.Store(current)
	p.writeFinal(current)
}

func (p *ProgressBar) shouldRender(current, total int64, percent int, now time.Time, force bool) bool {
	if force {
		return true
	}
	if p.flushInterval == 0 {
		return current != p.lastCurrent || total != p.lastTotal
	}
	if total > 0 && (percent != p.lastPercent || total != p.lastTotal) {
		return true
	}
	return now.Sub(p.lastRender) >= p.flushInterval
}

func (p *ProgressBar) render(current int64, force bool) {
	if p.writer == nil {
		return
	}
	now := time.Now()
	current = p.normalizeCurrent(current)
	total := p.Total()
	percent := p.percent(current)

	p.mu.Lock()
	if !p.shouldRender(current, total, percent, now, force) {
		p.mu.Unlock()
		return
	}
	meta := p.knownProgressMeta(current, percent, now, true)
	line := p.format(current, meta)
	p.lastRender = now
	p.lastPercent = percent
	p.lastCurrent = current
	p.lastTotal = total
	p.lastLine = line
	p.mu.Unlock()

	lineEnd := "\n"
	if p.interactive {
		lineEnd = "\r"
	}
	progressOutputMu.Lock()
	_, _ = p.writer.Write([]byte(line + lineEnd))
	progressOutputMu.Unlock()
}

func (p *ProgressBar) writeFinal(current int64) {
	if p.writer == nil {
		return
	}
	now := time.Now()
	current = p.normalizeCurrent(current)
	total := p.Total()
	percent := p.percent(current)

	p.mu.Lock()
	meta := p.knownProgressMeta(current, percent, now, true)
	line := p.format(current, meta)
	alreadyRendered := !p.lastRender.IsZero() && p.lastCurrent == current && p.lastPercent == percent && p.lastLine == line
	p.lastRender = now
	p.lastPercent = percent
	p.lastCurrent = current
	p.lastTotal = total
	p.lastLine = line
	p.mu.Unlock()

	progressOutputMu.Lock()
	if alreadyRendered {
		if p.interactive {
			_, _ = p.writer.Write([]byte("\n"))
		}
		progressOutputMu.Unlock()
		return
	}
	_, _ = p.writer.Write([]byte(line + "\n"))
	progressOutputMu.Unlock()
}

func (p *ProgressBar) format(current int64, meta string) string {
	termWidth, _ := terminalWidth(p.writer)
	core := ""
	if p.Total() > 0 {
		meta = shortenProgressMeta(meta, termWidth)
		barWidth := p.adaptiveBarWidth(meta)
		bar := ""
		if barWidth > 0 {
			bar = p.renderBar(barWidth, p.percent(current))
		}
		core = fitKnownProgressCore(bar, meta, termWidth)
	} else {
		core = fmt.Sprintf("[%c] %d", p.spinnerChar(current), current)
	}

	return fitProgressLine(p.prefix, core, p.suffix, termWidth)
}

func shortenProgressMeta(meta string, termWidth int) string {
	if termWidth <= 0 || stringDisplayWidth(meta) <= termWidth {
		return meta
	}

	shortened := strings.Replace(meta, " Elapsed ", " E ", 1)
	if stringDisplayWidth(shortened) <= termWidth {
		return shortened
	}

	shortened = strings.Replace(shortened, " ETA ", " T ", 1)
	if stringDisplayWidth(shortened) <= termWidth {
		return shortened
	}

	if idx := strings.Index(shortened, " E "); idx >= 0 {
		shortened = shortened[:idx]
		if stringDisplayWidth(shortened) <= termWidth {
			return shortened
		}
	}

	if idx := strings.Index(shortened, " T "); idx >= 0 {
		return shortened[:idx]
	}

	return shortened
}

func fitKnownProgressCore(bar, meta string, termWidth int) string {
	if termWidth <= 0 {
		if bar == "" {
			return meta
		}
		return bar + " " + meta
	}

	meta = truncateDisplayWidth(meta, termWidth)
	metaWidth := stringDisplayWidth(meta)
	if bar == "" || metaWidth >= termWidth {
		return meta
	}

	remaining := termWidth - metaWidth - 1
	if remaining < minAdaptiveBarWidth+2 {
		return meta
	}

	bar = truncateDisplayWidth(bar, remaining)
	if bar == "" {
		return meta
	}

	return bar + " " + meta
}

func (p *ProgressBar) knownProgressMeta(current int64, percent int, now time.Time, updateRate bool) string {
	total := p.Total()
	meta := fmt.Sprintf("%3d%% %d/%d", percent, current, total)
	if current <= 0 {
		return meta
	}
	elapsed := now.Sub(p.start)
	if elapsed < 0 {
		elapsed = 0
	}
	if updateRate {
		p.updateRateEstimate(current, now)
	}
	if percent < 100 && total > 0 {
		if eta, ok := p.estimatedETA(current, total, elapsed); ok {
			meta += fmt.Sprintf(" ETA %s", formatDuration(eta))
		}
	}
	meta += fmt.Sprintf(" Elapsed %s", formatDuration(elapsed))
	return meta
}

func (p *ProgressBar) updateRateEstimate(current int64, now time.Time) {
	deltaCount := current - p.lastSampleCurrent
	deltaTime := now.Sub(p.lastSampleTime)
	if deltaCount <= 0 || deltaTime <= 0 {
		return
	}
	rate := float64(deltaCount) / deltaTime.Seconds()
	if p.smoothedPerSecond <= 0 {
		p.smoothedPerSecond = rate
	} else {
		p.smoothedPerSecond = p.smoothedPerSecond*(1-etaSmoothingAlpha) + rate*etaSmoothingAlpha
	}
	p.lastSampleCurrent = current
	p.lastSampleTime = now
}

func (p *ProgressBar) estimatedETA(current, total int64, elapsed time.Duration) (time.Duration, bool) {
	remaining := total - current
	if remaining <= 0 {
		return 0, false
	}
	if p.smoothedPerSecond > 0 {
		seconds := float64(remaining) / p.smoothedPerSecond
		if seconds >= 0 {
			return time.Duration(seconds * float64(time.Second)), true
		}
	}
	if current > 0 && elapsed > 0 {
		seconds := float64(elapsed) * float64(remaining) / float64(current)
		return time.Duration(seconds), true
	}
	return 0, false
}

func (p *ProgressBar) adaptiveBarWidth(meta string) int {
	termWidth, ok := terminalWidth(p.writer)
	if !ok || termWidth <= 0 {
		return p.width
	}

	reserved := stringDisplayWidth(meta)
	return fitProgressBarWidth(p.width, reserved, termWidth)
}

func (p *ProgressBar) renderBar(width int, percent int) string {
	filled := int(float64(width) * float64(percent) / 100)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var b strings.Builder
	b.Grow(width + 2)
	b.WriteByte('[')
	for i := 0; i < width; i++ {
		switch {
		case i < filled:
			b.WriteByte(p.fill)
		case i == filled && percent < 100:
			b.WriteByte('>')
		default:
			b.WriteByte(p.empty)
		}
	}
	b.WriteByte(']')
	return b.String()
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d == 0 {
		return "0s"
	}
	return d.Truncate(time.Second).String()
}

func (p *ProgressBar) normalizeCurrent(current int64) int64 {
	total := p.Total()
	if current < 0 {
		return 0
	}
	if total > 0 && current > total {
		return total
	}
	return current
}

func normalizeTotal(total int64) int64 {
	if total < 0 {
		return 0
	}
	return total
}

func (p *ProgressBar) spinnerChar(current int64) rune {
	if len(p.spinChars) == 0 {
		return rune(defaultSpinnerChars[0])
	}
	if current <= 0 {
		return p.spinChars[0]
	}
	return p.spinChars[(current-1)%int64(len(p.spinChars))]
}

func (p *ProgressBar) percent(current int64) int {
	total := p.Total()
	if total <= 0 {
		return 0
	}
	ratio := float64(current) / float64(total)
	switch {
	case ratio < 0:
		ratio = 0
	case ratio > 1:
		ratio = 1
	}
	return int(ratio * 100)
}

func fitProgressBarWidth(preferred, reserved, termWidth int) int {
	if preferred <= 0 {
		preferred = defaultProgressWidth
	}
	available := termWidth - reserved - 1
	if available < minAdaptiveBarWidth+2 {
		return 0
	}
	if available < preferred+2 {
		preferred = available - 2
	}
	if preferred < minAdaptiveBarWidth {
		return 0
	}
	return preferred
}
