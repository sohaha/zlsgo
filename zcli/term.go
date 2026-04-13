package zcli

import (
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func terminalWidth(writer io.Writer) (int, bool) {
	file, ok := writer.(*os.File)
	if !ok || !fileIsTerminal(file) {
		return 0, false
	}
	if width, ok := fileTerminalWidth(file); ok && width > 0 {
		return width, true
	}
	if width, ok := envTerminalWidth(); ok {
		return width, true
	}
	return 0, false
}

func isTerminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	return fileIsTerminal(file)
}

func envTerminalWidth() (int, bool) {
	value := os.Getenv("COLUMNS")
	if value == "" {
		return 0, false
	}
	width, err := strconv.Atoi(value)
	if err != nil || width <= 0 {
		return 0, false
	}
	return width, true
}

func stringDisplayWidth(s string) int {
	width := 0
	for i := 0; i < len(s); {
		_, w, size := nextDisplayToken(s[i:])
		width += w
		i += size
	}
	return width
}

func fitProgressLine(prefix, core, suffix string, termWidth int) string {
	if termWidth <= 0 {
		return joinProgressLine(prefix, core, suffix)
	}

	core = truncateDisplayWidth(core, termWidth)
	coreWidth := stringDisplayWidth(core)
	if coreWidth >= termWidth {
		return core
	}

	remaining := termWidth - coreWidth
	if prefix != "" {
		prefix = truncateDisplayWidth(prefix, remaining-1)
		if prefix != "" {
			remaining -= stringDisplayWidth(prefix) + 1
		}
	}
	if suffix != "" {
		suffix = truncateDisplayWidth(suffix, remaining-1)
	}

	return joinProgressLine(prefix, core, suffix)
}

func joinProgressLine(prefix, core, suffix string) string {
	var b strings.Builder
	if prefix != "" {
		b.WriteString(prefix)
		if core != "" || suffix != "" {
			b.WriteByte(' ')
		}
	}
	if core != "" {
		b.WriteString(core)
		if suffix != "" {
			b.WriteByte(' ')
		}
	}
	if suffix != "" {
		b.WriteString(suffix)
	}
	return b.String()
}

func truncateDisplayWidth(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if stringDisplayWidth(s) <= max {
		return s
	}
	if max <= 3 {
		return strings.Repeat(".", max)
	}

	budget := max - 3
	var b strings.Builder
	width := 0
	for i := 0; i < len(s); {
		token, tokenWidth, size := nextDisplayToken(s[i:])
		if width+tokenWidth > budget {
			break
		}
		b.WriteString(token)
		width += tokenWidth
		i += size
	}
	b.WriteString("...")
	return b.String()
}

func nextDisplayToken(s string) (string, int, int) {
	if w, size := emojiClusterWidth(s); size > 0 {
		return s[:size], w, size
	}
	r, size := utf8.DecodeRuneInString(s)
	return s[:size], runeDisplayWidth(r), size
}

func emojiClusterWidth(s string) (int, int) {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size == 0 {
		return 0, 0
	}

	if isKeycapBase(r) {
		consumed := size
		next, nextSize := utf8.DecodeRuneInString(s[consumed:])
		if next == 0xfe0f {
			consumed += nextSize
			next, nextSize = utf8.DecodeRuneInString(s[consumed:])
		}
		if next == 0x20e3 {
			return 2, consumed + nextSize
		}
	}

	if !isEmojiBase(r) {
		return 0, 0
	}

	consumed := size
	if next, nextSize := utf8.DecodeRuneInString(s[consumed:]); next == 0xfe0f {
		consumed += nextSize
	}
	if next, nextSize := utf8.DecodeRuneInString(s[consumed:]); isEmojiModifier(next) {
		consumed += nextSize
	}

	for {
		next, nextSize := utf8.DecodeRuneInString(s[consumed:])
		if next != 0x200d {
			break
		}
		afterJoiner, afterJoinerSize := utf8.DecodeRuneInString(s[consumed+nextSize:])
		if !isEmojiBase(afterJoiner) {
			break
		}
		consumed += nextSize + afterJoinerSize
		if variation, variationSize := utf8.DecodeRuneInString(s[consumed:]); variation == 0xfe0f {
			consumed += variationSize
		}
		if modifier, modifierSize := utf8.DecodeRuneInString(s[consumed:]); isEmojiModifier(modifier) {
			consumed += modifierSize
		}
	}

	return 2, consumed
}

func isKeycapBase(r rune) bool {
	return (r >= '0' && r <= '9') || r == '#' || r == '*'
}

func isEmojiModifier(r rune) bool {
	return r >= 0x1f3fb && r <= 0x1f3ff
}

func isEmojiBase(r rune) bool {
	switch {
	case r >= 0x1f300 && r <= 0x1f5ff:
		return true
	case r >= 0x1f600 && r <= 0x1f64f:
		return true
	case r >= 0x1f680 && r <= 0x1f6ff:
		return true
	case r >= 0x1f700 && r <= 0x1f77f:
		return true
	case r >= 0x1f780 && r <= 0x1f7ff:
		return true
	case r >= 0x1f800 && r <= 0x1f8ff:
		return true
	case r >= 0x1f900 && r <= 0x1f9ff:
		return true
	case r >= 0x1fa70 && r <= 0x1faff:
		return true
	case r >= 0x2600 && r <= 0x26ff:
		return true
	case r >= 0x2700 && r <= 0x27bf:
		return true
	default:
		return false
	}
}

func runeDisplayWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 32 || (r >= 0x7f && r < 0xa0):
		return 0
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r), unicode.Is(unicode.Cf, r):
		return 0
	case unicode.Is(unicode.Han, r), unicode.Is(unicode.Hangul, r), unicode.Is(unicode.Hiragana, r), unicode.Is(unicode.Katakana, r):
		return 2
	case r >= 0x1100 && (r <= 0x115f ||
		r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f300 && r <= 0x1f64f) ||
		(r >= 0x1f900 && r <= 0x1f9ff) ||
		(r >= 0x20000 && r <= 0x3fffd)):
		return 2
	default:
		return 1
	}
}
