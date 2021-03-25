package zlog

import (
	"fmt"
	"os"
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestColor(t *testing.T) {
	T := zls.NewTest(t)
	testText := "ok"
	_ = os.Setenv("ConEmuANSI", "ON")
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
