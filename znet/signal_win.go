//go:build windows
// +build windows

package znet

import (
	"errors"
)

// Restart Restart
func (e *Engine) Restart() error {
	return errors.New("windows does not support")
}
