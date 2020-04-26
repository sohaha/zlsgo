package daemon

import (
	"testing"

	"fmt"

	"github.com/sohaha/zlsgo"
)

type ss struct {
	I int
}

func (p *ss) Start(s ServiceIfe) error {
	p.run()
	return nil
}

func (p *ss) run() {
	fmt.Println("run")
	p.I = p.I + 1
}

func (p *ss) Stop(s ServiceIfe) error {
	return nil
}

func TestDaemon(t *testing.T) {
	o := &ss{
		I: 1,
	}
	s, err := New(o, &Config{
		Name:   "zlsgo_daemon_test",
		Option: OptionData{"UserService": false},
	})
	if err != nil {
		return
	}
	t.Log(o.I)
	t.Log(err)
	_ = s.Install()
	err = s.Start()
	t.Log(err)
	_ = s.Stop()
	_ = s.Restart()
	t.Log(s.Status())
	_ = s.Uninstall()
	t.Log(s.String())
	t.Log(o.I)
}

func TestUtil(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(IsPermissionError(ErrNotAnAdministrator), IsPermissionError(ErrNotAnRootUser))
	_ = isSudo()
}

func TestOptionData(t *testing.T) {
	tt := zlsgo.NewTest(t)
	option := &OptionData{
		"string":     nil,
		"int":        nil,
		"bool":       nil,
		"float64":    nil,
		"funcSingle": nil,
	}
	tt.Equal("ss", option.string("string", "ss"))
	tt.Equal(11, option.int("int", 11))
	tt.Equal(true, option.bool("bool", true))
	tt.Equal(1.2, option.float64("float64", 1.2))
	option.funcSingle("funcSingle", func() {})
	option = &OptionData{
		"string":     "s",
		"int":        1,
		"bool":       true,
		"float64":    1.2,
		"funcSingle": func() {},
	}
	tt.Equal("s", option.string("string", "ss"))
	tt.Equal(1, option.int("int", 11))
	tt.Equal(true, option.bool("bool", true))
	tt.Equal(1.2, option.float64("float64", 1.1))
	option.funcSingle("funcSingle", func() {})
}
