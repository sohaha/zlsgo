package zlog

import (
	"fmt"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

// DisableColor DisableColor
var DisableColor = false

// Color Color
type Color int

// Format Format
type Format int

// Op Op
type Op int

const (
	// ColorBlack black
	ColorBlack Color = iota + 30
	// ColorRed gules
	ColorRed
	// ColorGreen green
	ColorGreen
	// ColorYellow yellow
	ColorYellow
	// ColorBlue blue
	ColorBlue
	// ColorMagenta magenta
	ColorMagenta
	// ColorCyan cyan
	ColorCyan
	// ColorWhite white
	ColorWhite
)

const (
	// ColorLightGrey light grey
	ColorLightGrey Color = iota + 90
	// ColorLightRed light red
	ColorLightRed
	// ColorLightGreen light green
	ColorLightGreen
	// ColorLightYellow light yellow
	ColorLightYellow
	// ColorLightBlue light blue
	ColorLightBlue
	// ColorLightMagenta light magenta
	ColorLightMagenta
	// ColorLightCyan lightcyan
	ColorLightCyan
	// ColorLightWhite light white
	ColorLightWhite
	// ColorDefault ColorDefault
	ColorDefault = 49
)
const (
	// OpReset Reset All Settings
	OpReset Op = iota
	// OpBold Bold
	OpBold
	// OpFuzzy Fuzzy (not all terminal emulators support it)
	OpFuzzy
	// OpItalic Italic (not all terminal emulators support it)
	OpItalic
	// OpUnderscore Underline
	OpUnderscore
	// OpBlink Twinkle
	OpBlink
	// OpFastBlink Fast scintillation (not widely supported)
	OpFastBlink
	// OpReverse Reversed Exchange Background and Foreground Colors
	OpReverse
	// OpConcealed Concealed
	OpConcealed
	// OpStrikethrough Deleted lines (not widely supported)
	OpStrikethrough
)

// OpTextWrap OpTextWrap
func OpTextWrap(op Op, text string) string {
	if !IsSupportColor() {
		return text
	}
	return fmt.Sprintf("\x1b[%dm%s", op, text)
}

// ColorBackgroundWrap ColorBackgroundWrap
func ColorBackgroundWrap(color Color, backgroundColor Color, text string) string {
	if !IsSupportColor() {
		return text
	}
	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", color, backgroundColor+10, text)
}

// OutAllColor OutAllColor
func OutAllColor() {
	all := zstring.Buffer()
	colors := GetAllColorText()
	for k, v := range colors {
		all.WriteString("\n\nBackground " + k + "\n")
		for ck, cv := range colors {
			if cv == v {
				continue
			}
			all.WriteString(ColorBackgroundWrap(cv, v, ck+" => "))
			all.WriteString(ColorBackgroundWrap(cv, v, OpTextWrap(OpBold, "Bold ")))
			all.WriteString(ColorBackgroundWrap(cv, v, OpTextWrap(OpUnderscore, "Under")))

			all.WriteString(ColorBackgroundWrap(cv, v, " | "))
		}
		all.WriteString("\n")
	}
	fmt.Println(all.String())
}

// GetAllColorText GetAllColorText
func GetAllColorText() map[string]Color {
	return map[string]Color{
		"ColorBlack":        ColorBlack,
		"ColorRed":          ColorRed,
		"ColorGreen":        ColorGreen,
		"ColorYellow":       ColorYellow,
		"ColorBlue":         ColorBlue,
		"ColorMagenta":      ColorMagenta,
		"ColorCyan":         ColorCyan,
		"ColorWhite":        ColorWhite,
		"ColorLightGrey":    ColorLightGrey,
		"ColorLightRed":     ColorLightRed,
		"ColorLightGreen":   ColorLightGreen,
		"ColorLightYellow":  ColorLightYellow,
		"ColorLightBlue":    ColorLightBlue,
		"ColorLightMagenta": ColorLightMagenta,
		"ColorLightCyan":    ColorLightCyan,
		"ColorLightWhite":   ColorLightWhite,
		"ColorDefault":      ColorDefault,
	}
}

// ColorTextWrap ColorTextWrap
func ColorTextWrap(color Color, text string) string {
	if !IsSupportColor() {
		return text
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

func isSupportColor() bool {
	return !DisableColor && (strings.Contains(os.Getenv("TERM"), "xterm") || os.Getenv("ConEmuANSI") == "ON" || os.Getenv("ANSICON") != "" || os.Getenv("ANSICON") != "" || strings.Contains(os.Getenv("TERM"), "256color"))
}
