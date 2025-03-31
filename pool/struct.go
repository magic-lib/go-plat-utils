package pool

import (
	"log"
	"reflect"
)

type ObjectBuilder[V any] interface {
	ID() string
	New() V
	Reset(v V) V
}

type BaseCreator[V any] struct {
	id string
}

func (b *BaseCreator[V]) New() (v V) {
	return v
}

func (b *BaseCreator[V]) ID() string {
	if b.id == "" {
		b.id = reflect.TypeOf(b.New()).String()
	}
	return b.id
}

// StructPool 基于struct的连接池
type StructPool[O ObjectBuilder[V], V any] struct {
	build O
}

func (s *StructPool[O, V]) InitPool(c O) error {
	s.build = c
	return registerPool[V](c)
}

// ClosePool 关闭连接池
func (s *StructPool[O, V]) ClosePool() error {
	closePool[V](s.build)
	return nil
}

// GetWithFunc 从连接池中获取一个连接，使用完毕后需要调用conn.Close()关闭连接
func (s *StructPool[O, V]) GetWithFunc(callFunc func(v V)) error {
	one := get[V](s.build)
	defer func(v V) {
		err := close[V](s.build, v)
		if err != nil {
			log.Println(err)
		}
	}(one)
	callFunc(one)
	return nil
}
