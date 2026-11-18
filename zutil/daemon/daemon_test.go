package daemon

import (
	"fmt"
	"testing"

	"github.com/sohaha/zlsgo"
)

type ss struct {
	I int
}

func (p *ss) Start(s ServiceIface) error {
	p.run()
	return nil
}

func (p *ss) run() {
	fmt.Println("run")
	p.I = p.I + 1
}

func (p *ss) Stop(s ServiceIface) error {
	return nil
}

func TestDaemon(t *testing.T) {
	o := &ss{
		I: 1,
	}
	s, err := New(o, &Config{
		Name:    "zlsgo_daemon_test",
		Options: map[string]interface{}{"UserService": false},
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

func TestStartDelay(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(0, startDelaySeconds(&Config{}))
	tt.Equal(0, startDelaySeconds(&Config{Options: map[string]interface{}{}}))
	tt.Equal(30, startDelaySeconds(&Config{Options: map[string]interface{}{"StartDelay": 30}}))
	tt.Equal(15, startDelaySeconds(&Config{Options: map[string]interface{}{"StartDelay": int64(15)}}))
	tt.Equal(10, startDelaySeconds(&Config{Options: map[string]interface{}{"StartDelay": "10"}}))
	tt.Equal(0, startDelaySeconds(&Config{Options: map[string]interface{}{"StartDelay": -5}}))
	tt.Equal(0, startDelaySeconds(&Config{Options: map[string]interface{}{"StartDelay": "abc"}}))
}
