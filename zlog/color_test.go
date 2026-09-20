package zlog

import (
	"fmt"
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestColor(t *testing.T) {
	oldDisableColor := DisableColor
	t.Cleanup(func() { DisableColor = oldDisableColor })
	DisableColor = false
	T := zls.NewTest(t)
	testText := "ok"
	t.Setenv("ConEmuANSI", "ON")
	bl := IsSupportColor()
	OutAllColor()
	if bl {
		T.Equal(fmt.Sprintf("\x1b[%dm%s\x1b[0m", ColorGreen, testText), ColorTextWrap(ColorGreen, testText))
	} else {
		T.Equal(testText, ColorTextWrap(ColorGreen, testText))
	}

	DisableColor = true
	bl = isSupportColor()
	T.Equal(false, bl)
	OutAllColor()
	T.Equal(testText, ColorTextWrap(ColorGreen, testText))
}

func TestTrimAnsi(t *testing.T) {
	tt := zls.NewTest(t)

	testText := "ok\x1b[31m"
	tt.Equal("ok", TrimAnsi(testText))

	testText = ColorTextWrap(ColorGreen, "ok")
	tt.Equal("ok", TrimAnsi(testText))
}
