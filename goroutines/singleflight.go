package goroutines

import (
	"errors"
	"sync"
)

// errFnPanicked fn 执行时 panic，等待中的调用方拿到该错误，而不是 (nil, nil)
var errFnPanicked = errors.New("goroutines: singleflight fn panicked")

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
		calls   map[string]*call
		results map[string]any // Once 成功执行后的结果
		lock    sync.Mutex
	}
)

// NewSingleFlight returns a SingleFlight.
func NewSingleFlight() SingleFlight {
	return &flightGroup{
		calls:   make(map[string]*call),
		results: make(map[string]any),
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
	// 原来读取的是进行中 call 的 val/err（锁外写入，存在数据竞争），且 call 完成后即被删除，实际从不缓存
	// 查缓存与登记 call 在同一把锁内完成；结果在删除 call 的同一把锁内写入，
	// 避免 call 已删除、结果尚未写入的窗口期里其他调用方再执行一次 fn
	g.lock.Lock()
	if cached, ok := g.results[key]; ok {
		g.lock.Unlock()
		return cached, nil
	}
	if c, ok := g.calls[key]; ok {
		g.lock.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.calls[key] = c
	g.lock.Unlock()

	g.doCall(c, key, fn, true)
	return c.val, c.err
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
	g.doCall(c, key, fn, false)
}

// doCall 执行 fn；saveResult 为 true 时，成功的结果在删除 call 的同一把锁内写入 results
func (g *flightGroup) doCall(c *call, key string, fn func() (any, error), saveResult bool) {
	normalReturn := false
	defer func() {
		if !normalReturn { // fn panic（panic 继续向上抛给执行方）
			c.val, c.err = nil, errFnPanicked
		}
		g.lock.Lock()
		if saveResult && c.err == nil && c.val != nil {
			g.results[key] = c.val
		}
		delete(g.calls, key)
		g.lock.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn()
	normalReturn = true
}
