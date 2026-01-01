package goroutines

import "sync"

type (
	SingleFlight interface {
		Do(key string, fn func() (any, error)) (any, bool, error)
		Once(key string, fn func() (any, error)) (any, error)
	}

	call struct {
		wg  sync.WaitGroup
		val any
		err error
	}

	flightGroup struct {
		calls map[string]*call
		lock  sync.Mutex
	}
)

// NewSingleFlight returns a SingleFlight.
func NewSingleFlight() SingleFlight {
	return &flightGroup{
		calls: make(map[string]*call),
	}
}

// Do 同时只有一个方法在执行
func (g *flightGroup) Do(key string, fn func() (any, error)) (val any, cached bool, err error) {
	c, done := g.createCall(key)
	if done {
		return c.val, done, c.err
	}

	g.makeCall(c, key, fn)
	return c.val, done, c.err
}

// Once 只执行一次，除非err!=nil
func (g *flightGroup) Once(key string, fn func() (any, error)) (val any, err error) {
	g.lock.Lock()
	if c, ok := g.calls[key]; ok {
		if c.val != nil && c.err == nil {
			g.lock.Unlock()
			return c.val, nil
		}
	}
	g.lock.Unlock()

	val, _, err = g.Do(key, fn)
	return val, err
}

func (g *flightGroup) createCall(key string) (c *call, done bool) {
	g.lock.Lock()
	if c, ok := g.calls[key]; ok {
		g.lock.Unlock()
		c.wg.Wait()
		return c, true
	}

	c = new(call)
	c.wg.Add(1)
	g.calls[key] = c
	g.lock.Unlock()

	return c, false
}

func (g *flightGroup) makeCall(c *call, key string, fn func() (any, error)) {
	defer func() {
		g.lock.Lock()
		delete(g.calls, key)
		g.lock.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn()
}
