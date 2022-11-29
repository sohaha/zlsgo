package znet

import (
	"io"
)

type Template interface {
	Load() error
	Render(io.Writer, string, interface{}, ...string) error
}

func (e *Engine) SetTemplate(v Template) {
	e.views = v
}
