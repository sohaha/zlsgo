package zshell

import "os/exec"

func wrapOptions(cmd *exec.Cmd, opt ...func(o *Options)) {
	o := Options{}
	for _, v := range opt {
		v(&o)
	}

	if o.Dir != "" {
		cmd.Dir = o.Dir
	} else if Dir != "" {
		cmd.Dir = Dir
	}

	if o.Env != nil {
		cmd.Env = append(cmd.Env, o.Env...)
	} else if Env != nil {
		cmd.Env = append(cmd.Env, Env...)
	}
}
