package zerror

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/sohaha/zlsgo/zutil"
)

// Error returns the msg of e
func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}
	return e.err.Error()
}

// Unwrap returns err inside
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.wrapErr
}

// Format formats the frame according to the fmt.Formatter interface.
//
// %v, %s   : Print all the error string;
// %-v, %-s : Print current level error string;
// %+s      : Print full stack error list;
// %+v      : Print the error string and full stack error list;
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = io.WriteString(s, e.Error()+"\n"+e.Stack())
		default:
			_, _ = io.WriteString(s, e.Error())
		}
	}
}

// Stack returns the stack callers as string.
// It returns an empty string if the `err` does not support stacks.
func (e *Error) Stack() string {
	if e == nil {
		return ""
	}
	loop, i, buffer := error(e), 1, zutil.GetBuff()
	defer zutil.PutBuff(buffer)
	for {
		if loop == nil {
			break
		}
		buffer.WriteString(fmt.Sprintf("%d. %-v\n", i, loop))
		i++
		e, ok := loop.(*Error)
		if ok {
			formatSubStack(e.stack, buffer)
			if e.wrapErr != nil {
				if e, ok := e.wrapErr.(*Error); ok {
					loop = e
				} else {
					buffer.WriteString(fmt.Sprintf("%d. %s\n", i, e.err.Error()))
					i++
					break
				}
			} else {
				break
			}
		}
	}
	return buffer.String()
}

// formatSubStack formats the stack for error.
func formatSubStack(st zutil.Stack, buffer *bytes.Buffer) {
	if st == nil {
		return
	}
	index := 1
	space := "  "
	st.Format(func(fn *runtime.Func, file string, line int) bool {
		if strings.HasSuffix(file, "zerror/error.go") {
			return true
		}
		if strings.Contains(file, "<") {
			return true
		}
		if goROOT != "" && strings.HasPrefix(file, goROOT) {
			return true
		}
		if index > 9 {
			space = " "
		}
		buffer.WriteString(fmt.Sprintf(
			"   %d).%s%s\n    \t%s:%d\n",
			index, space, fn.Name(), file, line,
		))
		index++
		return true
	})
}
