package zlog

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/sohaha/zlsgo/zreflect"
)

type indentWriter struct {
	w   io.Writer
	bol bool
	pre [][]byte
	sel int
	off int
}

func NewIndentWriter(w io.Writer, pre ...[]byte) io.Writer {
	return &indentWriter{
		w:   w,
		pre: pre,
		bol: true,
	}
}

func (w *indentWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if w.bol {
			var i int
			i, err = w.w.Write(w.pre[w.sel][w.off:])
			w.off += i
			if err != nil {
				return n, err
			}
		}
		_, err = w.w.Write([]byte{c})
		if err != nil {
			return n, err
		}
		n++
		w.bol = c == '\n'
		if w.bol {
			w.off = 0
			if w.sel < len(w.pre)-1 {
				w.sel++
			}
		}
	}
	return n, nil
}

func argName(arg ast.Expr) string {
	name := ""

	switch a := arg.(type) {
	case *ast.Ident:
		switch {
		case a.Obj == nil:
			name = a.Name
		case a.Obj.Kind == ast.Var, a.Obj.Kind == ast.Con:
			name = a.Obj.Name
		}
	case *ast.BinaryExpr,
		*ast.CallExpr,
		*ast.IndexExpr,
		*ast.KeyValueExpr,
		*ast.ParenExpr,
		*ast.SelectorExpr,
		*ast.SliceExpr,
		*ast.TypeAssertExpr,
		*ast.UnaryExpr:
		name = exprToString(arg)
	}

	return name
}

func argNames(filename string, line int) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", filename, err)
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, is := n.(*ast.CallExpr)
		if !is {
			return true
		}
		if fset.Position(call.End()).Line != line {
			return true
		}
		for _, arg := range call.Args {
			names = append(names, argName(arg))
		}
		return true
	})

	return names, nil
}

func exprToString(arg ast.Expr) string {
	var buf strings.Builder
	fset := token.NewFileSet()
	if err := printer.Fprint(&buf, fset, arg); err != nil {
		return ""
	}
	return strings.Replace(buf.String(), "\t", "    ", -1)
}

func (fo formatter) String() string {
	return fmt.Sprint(fo.v.Interface())
}

func (fo formatter) Format(f fmt.State, c rune) {
	if fo.force || c == 'v' && f.Flag('#') && f.Flag(' ') {
		w := tabwriter.NewWriter(f, 4, 4, 1, ' ', 0)
		p := &zprinter{tw: w, Writer: w, visited: make(map[visit]int)}
		p.printValue(fo.v, true, fo.quote)
		_ = w.Flush()
		return
	}
	fo.passThrough(f, c)
}

func (fo formatter) passThrough(f fmt.State, c rune) {
	s := "%"
	for i := 0; i < 128; i++ {
		if f.Flag(i) {
			s += strconv.FormatInt(int64(i), 10)
		}
	}
	if w, ok := f.Width(); ok {
		s += fmt.Sprintf("%d", w)
	}
	if p, ok := f.Precision(); ok {
		s += fmt.Sprintf(".%d", p)
	}
	s += string(c)
	_, _ = fmt.Fprintf(f, s, fo.v.Interface())
}

func (p *zprinter) indent() *zprinter {
	q := *p
	q.tw = tabwriter.NewWriter(p.Writer, 4, 4, 1, ' ', 0)
	q.Writer = NewIndentWriter(q.tw, []byte{'\t'})
	return &q
}

func (p *zprinter) printInline(v reflect.Value, x interface{}, showType bool) {
	if showType {
		_, _ = io.WriteString(p, v.Type().String())
		_, _ = fmt.Fprintf(p, "(%+v)", x)
	} else {
		_, _ = fmt.Fprintf(p, "%+v", x)
	}
}

func (p *zprinter) printValue(v reflect.Value, showType, quote bool) {
	if p.depth > 10 {
		_, _ = io.WriteString(p, "!%v(DEPTH EXCEEDED)")
		return
	}
	switch v.Kind() {
	case reflect.Bool:
		p.printInline(v, v.Bool(), showType)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.printInline(v, v.Int(), showType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.printInline(v, v.Uint(), showType)
	case reflect.Float32, reflect.Float64:
		p.printInline(v, v.Float(), showType)
	case reflect.Complex64, reflect.Complex128:
		_, _ = fmt.Fprintf(p, "%#v", v.Complex())
	case reflect.String:
		p.fmtString(v.String(), quote)
	case reflect.Map:
		t := v.Type()
		if showType {
			_, _ = io.WriteString(p, t.String())
		}
		writeByte(p, '{')
		if zreflect.Nonzero(v) {
			expand := !zreflect.CanInline(v.Type())
			pp := p
			if expand {
				writeByte(p, '\n')
				pp = p.indent()
			}
			keys := v.MapKeys()
			for i := 0; i < v.Len(); i++ {
				k := keys[i]
				mv := v.MapIndex(k)
				pp.printValue(k, false, true)
				writeByte(pp, ':')
				if expand {
					writeByte(pp, '\t')
				}
				showTypeInStruct := t.Elem().Kind() == reflect.Interface
				pp.printValue(mv, showTypeInStruct, true)
				if expand {
					_, _ = io.WriteString(pp, ",\n")
				} else if i < v.Len()-1 {
					_, _ = io.WriteString(pp, ", ")
				}
			}
			if expand {
				_ = pp.tw.Flush()
			}
		}
		writeByte(p, '}')
	case reflect.Struct:
		t := v.Type()
		if v.CanAddr() {
			addr := v.UnsafeAddr()
			vis := visit{addr, t}
			if vd, ok := p.visited[vis]; ok && vd < p.depth {
				p.fmtString(t.String()+"{(CYCLIC REFERENCE)}", false)
				break // don't print v again
			}
			p.visited[vis] = p.depth
		}

		if showType {
			_, _ = io.WriteString(p, t.String())
		}
		writeByte(p, '{')
		if zreflect.Nonzero(v) {
			expand := !zreflect.CanInline(v.Type())
			pp := p
			if expand {
				writeByte(p, '\n')
				pp = p.indent()
			}
			for i := 0; i < v.NumField(); i++ {
				showTypeInStruct := true
				if f := t.Field(i); f.Name != "" {
					_, _ = io.WriteString(pp, f.Name)
					writeByte(pp, ':')
					if expand {
						writeByte(pp, '\t')
					}
					showTypeInStruct = zreflect.IsLabelType(f.Type)
				}
				pp.printValue(zreflect.GetInterfaceField(v, i), showTypeInStruct, true)
				if expand {
					_, _ = io.WriteString(pp, ",\n")
				} else if i < v.NumField()-1 {
					_, _ = io.WriteString(pp, ", ")
				}
			}
			if expand {
				_ = pp.tw.Flush()
			}
		}
		writeByte(p, '}')
	case reflect.Interface:
		switch e := v.Elem(); {
		case e.Kind() == reflect.Invalid:
			_, _ = io.WriteString(p, "nil")
		case e.IsValid():
			pp := *p
			pp.depth++
			pp.printValue(e, showType, true)
		default:
			_, _ = io.WriteString(p, v.Type().String())
			_, _ = io.WriteString(p, "(nil)")
		}
	case reflect.Array, reflect.Slice:
		t := v.Type()
		if showType {
			_, _ = io.WriteString(p, t.String())
		}
		if v.Kind() == reflect.Slice && v.IsNil() && showType {
			_, _ = io.WriteString(p, "(nil)")
			break
		}
		if v.Kind() == reflect.Slice && v.IsNil() {
			_, _ = io.WriteString(p, "nil")
			break
		}
		writeByte(p, '{')
		expand := !zreflect.CanInline(v.Type())
		pp := p
		if expand {
			writeByte(p, '\n')
			pp = p.indent()
		}
		for i := 0; i < v.Len(); i++ {
			showTypeInSlice := t.Elem().Kind() == reflect.Interface
			pp.printValue(v.Index(i), showTypeInSlice, true)
			if expand {
				_, _ = io.WriteString(pp, ",\n")
			} else if i < v.Len()-1 {
				_, _ = io.WriteString(pp, ", ")
			}
		}
		if expand {
			_ = pp.tw.Flush()
		}
		writeByte(p, '}')
	case reflect.Ptr:
		e := v.Elem()
		if !e.IsValid() {
			writeByte(p, '(')
			_, _ = io.WriteString(p, v.Type().String())
			_, _ = io.WriteString(p, ")(nil)")
		} else {
			pp := *p
			pp.depth++
			writeByte(pp, '&')
			pp.printValue(e, true, true)
		}
	case reflect.Chan:
		x := v.Pointer()
		if showType {
			writeByte(p, '(')
			_, _ = io.WriteString(p, v.Type().String())
			_, _ = fmt.Fprintf(p, ")(%#v)", x)
		} else {
			_, _ = fmt.Fprintf(p, "%#v", x)
		}
	case reflect.Func:
		_, _ = io.WriteString(p, v.Type().String())
		_, _ = io.WriteString(p, " {...}")
	case reflect.UnsafePointer:
		p.printInline(v, v.Pointer(), showType)
	case reflect.Invalid:
		_, _ = io.WriteString(p, "nil")
	}
}

func (p *zprinter) fmtString(s string, quote bool) {
	if quote {
		s = strconv.Quote(s)
	}
	_, _ = io.WriteString(p, s)
}
