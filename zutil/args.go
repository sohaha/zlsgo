package zutil

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// Args provides a flexible argument handling system primarily designed for SQL query building.
// It supports named arguments, positional arguments, and custom compilation handlers.
// This allows for dynamic parameter substitution in query strings.
type Args struct {
	namedArgs      map[string]int
	sqlNamedArgs   map[string]int
	compileHandler ArgsCompileHandler
	args           []argsArr
	onlyNamed      bool
}

// argsArr represents a single argument with an optional transformation function.
type argsArr struct {
	// Fn is an optional function that can transform the argument value
	Fn func(k string) interface{}
	// Arg is the actual argument value
	Arg interface{}
}

// ArgsOpt is a function type for configuring an Args instance using the functional options pattern.
type ArgsOpt func(*Args)

// ArgsCompileHandler is a function type for custom argument compilation.
// It allows for custom handling of how arguments are formatted and added to the query.
type ArgsCompileHandler func(buf *bytes.Buffer, values []interface{}, arg interface{}) ([]interface{}, bool)

const maxPredefinedArgs = 64

var predefinedArgs []string

func init() {
	predefinedArgs = make([]string, 0, maxPredefinedArgs)
	for i := 0; i < maxPredefinedArgs; i++ {
		predefinedArgs = append(predefinedArgs, fmt.Sprintf("$%v", i))
	}
}

// WithOnlyNamed returns an option function that configures Args to only use named parameters.
// When this option is set, positional parameters (like $1, $2) will not be processed.
func WithOnlyNamed() func(args *Args) {
	return func(args *Args) {
		args.onlyNamed = true
	}
}

// WithCompileHandler returns an option function that sets a custom compile handler for Args.
// The compile handler determines how arguments are formatted and added to the query.
func WithCompileHandler(fn ArgsCompileHandler) func(args *Args) {
	return func(args *Args) {
		args.compileHandler = fn
	}
}

// NewArgs creates a new Args instance with the provided options.
// This is the entry point for using the argument handling system.
func NewArgs(opt ...ArgsOpt) *Args {
	args := &Args{}
	for _, o := range opt {
		o(args)
	}
	return args
}

// Var adds an argument to the Args instance and returns a placeholder string.
// The placeholder can be used in a query string and will be replaced with the
// actual argument value during compilation.
func (args *Args) Var(arg interface{}) string {
	idx := args.add(arg, nil)
	if idx < maxPredefinedArgs {
		return predefinedArgs[idx]
	}
	return fmt.Sprintf("$%v", idx)
}

// add adds an argument to the Args instance and returns its index.
// This is an internal method used by Var and other methods.
func (args *Args) add(arg interface{}, fn func(k string) interface{}) int {
	idx := len(args.args)

	switch a := arg.(type) {
	case namedArgs:
		if args.namedArgs == nil {
			args.namedArgs = map[string]int{}
		}
		if p, ok := args.namedArgs[a.name]; ok {
			arg = args.args[p]
			break
		}
		arg := a.arg
		switch v := a.arg.(type) {
		default:
			idx = args.add(arg, nil)
		case func() interface{}:
			idx = args.add(arg, func(_ string) interface{} { return v() })
		case func(k string) interface{}:
			idx = args.add(arg, v)
		}

		args.namedArgs[a.name] = idx
		return idx
	case sql.NamedArg:
		if args.sqlNamedArgs == nil {
			args.sqlNamedArgs = map[string]int{}
		}
		if p, ok := args.sqlNamedArgs[a.Name]; ok {
			arg = args.args[p]
			break
		}

		args.sqlNamedArgs[a.Name] = idx
	}

	args.args = append(args.args, argsArr{Arg: arg, Fn: fn})
	return idx
}

// CompileString compiles a format string with the arguments and returns the result as a string.
// This is useful for generating human-readable representations of queries.
func (args *Args) CompileString(format string, initialValue ...interface{}) string {
	old := args.compileHandler
	args.compileHandler = func(buf *bytes.Buffer, values []interface{}, arg interface{}) ([]interface{}, bool) {
		switch v := arg.(type) {
		case string:
			buf.WriteString(v)
		case sql.NamedArg:
			buf.WriteString(ztype.ToString(v.Value))
		default:
			val := ztype.ToString(v)
			buf.WriteString(val)
		}
		return values, true
	}
	defer func() {
		if old != nil {
			args.compileHandler = old
		}
	}()
	query, _ := args.Compile(format, initialValue...)

	return query
}

// Compile processes a format string with placeholders and returns the compiled query
// and a slice of argument values. This is the main method for generating SQL queries
// with proper parameter substitution.
func (args *Args) Compile(format string, initialValue ...interface{}) (query string, values []interface{}) {
	buf := GetBuff(256)
	idx := strings.IndexRune(format, '$')
	offset := 0
	values = initialValue

	for idx >= 0 && len(format) > 0 {
		if idx > 0 {
			buf.WriteString(format[:idx])
		}

		format = format[idx+1:]
		if len(format) == 0 {
			buf.WriteRune('$')
			break
		}

		if r := format[0]; r == '$' {
			buf.WriteRune('$')
			format = format[1:]
		} else if r == '{' {
			format, values = args.compileNamed(buf, format, values)
		} else if !args.onlyNamed && '0' <= r && r <= '9' {
			format, values, offset = args.compileDigits(buf, format, values, offset)
		} else if !args.onlyNamed && r == '?' {
			format, values, offset = args.compileSuccessive(buf, format[1:], values, offset, "")
		} else {
			buf.WriteRune('$')
		}

		idx = strings.IndexRune(format, '$')
	}

	if len(format) > 0 {
		buf.WriteString(format)
	}

	query = buf.String()

	PutBuff(buf)

	if len(args.sqlNamedArgs) > 0 {
		ints := make([]int, 0, len(args.sqlNamedArgs))
		for _, p := range args.sqlNamedArgs {
			ints = append(ints, p)
		}
		sort.Ints(ints)

		for _, i := range ints {
			values = append(values, args.args[i].Arg)
		}
	}

	return
}

// compileNamed compiles a named parameter in the format string.
func (args *Args) compileNamed(buf *bytes.Buffer, format string, values []interface{}) (string, []interface{}) {
	i := 1
	for ; i < len(format) && format[i] != '}'; i++ {
	}
	if i == len(format) {
		return format, values
	}

	name := format[1:i]
	format = format[i+1:]

	if p, ok := args.namedArgs[name]; ok {
		format, values, _ = args.compileSuccessive(buf, format, values, p, "")
	} else if strings.IndexRune(name, '.') > 0 {
		for n := range args.namedArgs {
			if zstring.Match(name, n) {
				p := args.namedArgs[n]
				format, values, _ = args.compileSuccessive(buf, format, values, p, name)
			}
		}
	}

	return format, values
}

// compileDigits compiles a positional parameter in the format string.
func (args *Args) compileDigits(buf *bytes.Buffer, format string, values []interface{}, offset int) (string, []interface{}, int) {
	i := 1
	for ; i < len(format) && '0' <= format[i] && format[i] <= '9'; i++ {
	}

	digits := format[:i]
	format = format[i:]

	if pointer, err := strconv.Atoi(digits); err == nil {
		return args.compileSuccessive(buf, format, values, pointer, "")
	}

	return format, values, offset
}

// compileSuccessive compiles a successive parameter in the format string.
func (args *Args) compileSuccessive(buf *bytes.Buffer, format string, values []interface{}, offset int, name string) (string, []interface{}, int) {
	if offset >= len(args.args) {
		return format, values, offset
	}

	arg := args.args[offset]
	if arg.Fn != nil {
		values = args.CompileArg(buf, values, arg.Fn(name))
	} else {
		values = args.CompileArg(buf, values, arg.Arg)
	}

	return format, values, offset + 1
}

// CompileArg compiles a single argument and appends it to the values slice.
func (args *Args) CompileArg(buf *bytes.Buffer, values []interface{}, arg interface{}) []interface{} {
	if args.compileHandler != nil {
		if values, ok := args.compileHandler(buf, values, arg); ok {
			return values
		}
	}
	switch a := arg.(type) {
	case sql.NamedArg:
		buf.WriteRune('@')
		buf.WriteString(a.Name)
	default:
		buf.WriteRune('?')
		values = append(values, arg)
	}

	return values
}
