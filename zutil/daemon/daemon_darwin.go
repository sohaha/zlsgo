package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/sohaha/zlsgo/zerror"
)

type (
	darwinSystem         struct{}
	darwinLaunchdService struct {
		i Iface
		*Config
		userService bool
	}
)

const version = "darwin-launchd"

var interactive = false

func (darwinSystem) String() string {
	return version
}

func (darwinSystem) Detect() bool {
	return true
}

func (darwinSystem) Interactive() bool {
	return interactive
}

func (darwinSystem) New(i Iface, c *Config) (s ServiceIface, err error) {
	userService := optionUserServiceDefault
	if s, ok := c.Options[optionUserService]; ok {
		userService, _ = s.(bool)
	}
	s = &darwinLaunchdService{
		i:           i,
		Config:      c,
		userService: userService,
	}
	if !userService {
		err = isSudo()
	}
	return s, err
}

func init() {
	var err error
	chooseSystem(darwinSystem{})
	interactive, err = isInteractive()
	zerror.Panic(err)
}

func isInteractive() (bool, error) {
	return os.Getppid() != 1, nil
}

func (s *darwinLaunchdService) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}

func (s *darwinLaunchdService) getHomeDir() (string, error) {
	u, err := user.Current()
	if err == nil {
		return u.HomeDir, nil
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "", errors.New("user home directory not found")
	}
	return homeDir, nil
}

func (s *darwinLaunchdService) getServiceFilePath() (string, error) {
	if s.userService {
		homeDir, err := s.getHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir + "/Library/LaunchAgents/" + s.Name + ".plist", nil
	}
	return "/Library/LaunchDaemons/" + s.Name + ".plist", nil
}

func (s *darwinLaunchdService) Install() error {
	confPath, err := s.getServiceFilePath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("init already exists: %s", confPath)
	}

	if s.userService {
		// ~/Library/LaunchAgents exists
		err = os.MkdirAll(filepath.Dir(confPath), 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	keepAlive := optionKeepAliveDefault
	if v, ok := s.Options[optionKeepAlive]; ok {
		keepAlive, _ = v.(bool)
	}
	load := optionRunAtLoadDefault
	if v, ok := s.Options[optionRunAtLoad]; ok {
		load, _ = v.(bool)
	}
	sessionCreate := optionSessionCreateDefault
	if v, ok := s.Options[optionSessionCreate]; ok {
		sessionCreate, _ = v.(bool)
	}

	path := s.execPath()
	to := &struct {
		*Config
		Path string

		KeepAlive, RunAtLoad bool
		SessionCreate        bool
	}{
		Config:        s.Config,
		Path:          path,
		KeepAlive:     keepAlive,
		RunAtLoad:     load,
		SessionCreate: sessionCreate,
	}

	functions := template.FuncMap{
		"bool": func(v bool) string {
			if v {
				return "true"
			}
			return "false"
		},
	}
	t := template.Must(template.New("launchdConfig").Funcs(functions).Parse(launchdConfig))
	return t.Execute(f, to)
}

func (s *darwinLaunchdService) Uninstall() error {
	var (
		err      error
		confPath string
	)
	if err = s.Stop(); err != nil {
		return err
	}
	if confPath, err = s.getServiceFilePath(); err != nil {
		return err
	}
	return os.Remove(confPath)
}

func (s *darwinLaunchdService) Start() error {
	confPath, err := s.getServiceFilePath()
	if err != nil {
		return err
	}
	err = run("launchctl", "load", confPath)

	return err
}

func (s *darwinLaunchdService) Stop() error {
	confPath, err := s.getServiceFilePath()
	if err != nil {
		return err
	}
	_ = run("launchctl", "stop", confPath)
	for {
		err = run("launchctl", "unload", confPath)
		if err == nil || (strings.Contains(err.Error(), "Could not find specified service") || !strings.Contains(err.Error(), "Operation now in progress")) {
			err = nil
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func (s *darwinLaunchdService) Status() string {
	res, _ := runGrep(s.Name+"$", "launchctl", "list")
	if res != "" {
		return "Running"
	}
	return "Stop"
}

func (s *darwinLaunchdService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func (s *darwinLaunchdService) Run() error {
	err := s.i.Start(s)
	if err != nil {
		return err
	}
	runWait := func() {
		<-SingleKillSignal()
	}
	if v, ok := s.Options[optionRunWait]; ok {
		runWait, _ = v.(func())
	}

	runWait()
	return s.i.Stop(s)
}

var launchdConfig = `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN"
"http://www.apple.com/DTDs/PropertyList-1.0.dtd" >
<plist version='1.0'>
<dict>
<key>Label</key><string>{{html .Name}}</string>
<key>ProgramArguments</key>
<array>
        <string>{{html .Path}}</string>
{{range .Config.Arguments}}
        <string>{{html .}}</string>
{{end}}
</array>
{{if .UserName}}<key>UserName</key><string>{{html .UserName}}</string>{{end}}
{{if .RootDir}}<key>RootDirectory</key><string>{{html .RootDir}}</string>{{end}}
{{if .WorkingDir}}<key>WorkingDirectory</key><string>{{html .WorkingDir}}</string>{{end}}
<key>SessionCreate</key><{{bool .SessionCreate}}/>
<key>KeepAlive</key><{{bool .KeepAlive}}/>
<key>RunAtLoad</key><{{bool .RunAtLoad}}/>
<key>Disabled</key><false/>
</dict>
</plist>
`

// <key>StandardOutPath</key>
// <string>/tmp/zlsgo/{{html .Name}}.log</string>
// <key>StandardErrorPath</key>
// <string>/tmp/zlsgo/{{html .Name}}.err</string>
