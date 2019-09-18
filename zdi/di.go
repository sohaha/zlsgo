package zdi

import (
	"errors"
	"sync"
)

// IfeDi IfeDi
type IfeDi interface {
	Remove(name string)
	Exist(name string) bool
	Make(name string) interface{}
	Bind(name string, v interface{})
	SoftMake(name string, v interface{}) (err error)
}

// Di Di
type Di struct {
	store map[string]interface{}
	mutex sync.RWMutex
}

// New create a di instance
func New() IfeDi {
	d := new(Di)
	d.store = make(map[string]interface{})
	return d
}

// SoftMake Register if the container does not exist
func (d *Di) SoftMake(name string, v interface{}) (err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if _, ok := d.store[name]; ok {
		return errors.New("container is set")
	}

	d.store[name] = v
	return
}

// Bind Registration container
func (d *Di) Bind(name string, v interface{}) {
	d.mutex.Lock()
	d.store[name] = v
	d.mutex.Unlock()
}

// Make Make the specified container, return nil if it does not exist
func (d *Di) Make(name string) interface{} {
	d.mutex.RLock()
	v := d.store[name]
	d.mutex.RUnlock()
	return v
}

// Exist whether the container exists
func (d *Di) Exist(name string) bool {
	d.mutex.RLock()
	_, ok := d.store[name]
	d.mutex.RUnlock()
	return ok
}

// Remove Unbind container
func (d *Di) Remove(name string) {
	d.mutex.Lock()
	delete(d.store, name)
	d.mutex.Unlock()
}
