package daemon

import (
	"fmt"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
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
		Option: zarray.DefData{"UserService": false},
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
