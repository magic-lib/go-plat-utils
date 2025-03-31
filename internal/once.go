package internal

import (
	"sync"
	"sync/atomic"
)

type Once struct {
	done atomic.Uint32
	m    sync.Mutex
}

// Do calls the function f if and only if Do is being called for the
func (o *Once) Do(f func() error) error {
	if o.done.Load() == 0 {
		// Outlined slow-path to allow inlining of the fast-path.
		return o.doSlow(f)
	}
	return nil
}

func (o *Once) doSlow(f func() error) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done.Load() == 0 {
		err := f()
		if err == nil {
			o.done.Store(1)
		}
		return err
	}
	return nil
}
