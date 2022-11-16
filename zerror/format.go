package zerror

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/sohaha/zlsgo/zutil"
)

// Error returns msg
func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}
	if e.inner && e.wrapErr != nil {
		return e.wrapErr.Error()
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

// Format formats the frame according to the fmt.Formatter interface
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			tip := strings.Join(UnwrapErrors(e), ": ")
			_, _ = io.WriteString(s, tip+"\n"+e.Stack())
		default:
			_, _ = io.WriteString(s, e.Error())
		}
	case 's':
		_, _ = io.WriteString(s, strings.Join(UnwrapErrors(e), ": "))
	}
}

// Stack returns the stack callers as string
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
		e, ok := loop.(*Error)
		if ok {
			if e.stack != nil {
				if e.err != nil {
					buffer.WriteString(fmt.Sprintf("%d. %-v\n", i, e.err))
				} else {
					buffer.WriteString(fmt.Sprintf("%d. %-v\n", i, e))
				}
				i++
				formatSubStack(e.stack, buffer)
			}
			if e.wrapErr != nil {
				if en, ok := e.wrapErr.(*Error); ok {
					loop = en
				} else {
					loop = e.wrapErr
					if loop == nil {
						break
					}
					buffer.WriteString(fmt.Sprintf("%d. %s\n", i, loop.Error()))
					break
				}
			} else {
				break
			}
		}
	}
	return buffer.String()
}

// formatSubStack formats the stack for error
func formatSubStack(st zutil.Stack, buffer *bytes.Buffer) {
	if st == nil {
		return
	}
	index := 1
	space := "  "
	st.Format(func(fn *runtime.Func, file string, line int) bool {
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
