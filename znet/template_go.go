package znet

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
)

type htmlEngine struct {
	log       *zlog.Logger
	funcmap   map[string]interface{}
	Templates *template.Template
	directory string
	options   TemplateOptions
	mutex     sync.RWMutex
	loaded    bool
}

type TemplateOptions struct {
	Extension  string
	Layout     string
	DelimLeft  string
	DelimRight string
	Reload     bool
	Debug      bool
}

func getTemplateOptions(debug bool, opt ...func(o *TemplateOptions)) TemplateOptions {
	o := TemplateOptions{
		Extension:  ".html",
		DelimLeft:  "{{",
		DelimRight: "}}",
		Layout:     "slot",
	}
	if debug {
		o.Debug = true
		o.Reload = true
	}
	for _, f := range opt {
		f(&o)
	}
	return o
}

var _ Template = &htmlEngine{}

func newGoTemplate(e *Engine, directory string, opt ...func(o *TemplateOptions)) *htmlEngine {
	h := &htmlEngine{
		directory: directory,
		funcmap:   make(map[string]interface{}),
	}
	if e != nil {
		h.log = e.Log
		h.options = getTemplateOptions(e.IsDebug(), opt...)
	} else {
		h.log = zlog.New()
		h.log.ResetFlags(zlog.BitLevel)
		h.options = getTemplateOptions(true, opt...)
	}
	return h
}

func (e *htmlEngine) AddFunc(name string, fn interface{}) *htmlEngine {
	e.mutex.Lock()
	e.funcmap[name] = fn
	e.mutex.Unlock()
	return e
}

func (e *htmlEngine) SetFuncMap(m map[string]interface{}) *htmlEngine {
	e.funcmap = m
	return e
}

func (e *htmlEngine) Load() error {
	if e.loaded && !e.options.Reload {
		return nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Templates = template.New(e.directory)
	e.Templates.Delims(e.options.DelimLeft, e.options.DelimRight)
	e.Templates.Funcs(e.funcmap)

	total := 0
	tip := zstring.Buffer()
	err := filepath.Walk(e.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil || info.IsDir() {
			return nil
		}

		if len(e.options.Extension) >= len(path) || path[len(path)-len(e.options.Extension):] != e.options.Extension {
			return nil
		}

		rel, err := filepath.Rel(e.directory, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		buf, err := zfile.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = e.Templates.New(name).Parse(string(buf))
		if err == nil {
			if e.options.Debug {
				total++
				tip.WriteString("\t    - " + name + "\n")
			}
		}

		return err
	})
	if err == nil && !e.loaded && e.options.Debug {
		e.log.Debugf("Loaded HTML Templates (%d): \n%s", total, tip.String())
	}

	e.loaded = true
	return err
}

func (e *htmlEngine) Render(out io.Writer, template string, data interface{}, layout ...string) error {
	if !e.loaded || e.options.Reload {
		if err := e.Load(); err != nil {
			return err
		}
	}

	tmpl := e.Templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("template %s does not exist", template)
	}

	if len(layout) > 0 && layout[0] != "" {
		lay := e.Templates.Lookup(layout[0])
		if lay == nil {
			return fmt.Errorf("layout %s does not exist", layout[0])
		}
		e.mutex.Lock()
		defer e.mutex.Unlock()
		lay.Funcs(map[string]interface{}{
			e.options.Layout: func() error {
				return tmpl.Execute(out, data)
			},
		})
		return lay.Execute(out, data)
	}

	return tmpl.Execute(out, data)
}
