package pool

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

// RedisPool 基于redigo的redis连接池
type RedisPool struct {
	redisPool *redis.Pool
}

// InitPool 新建一个redis连接池
func (p *RedisPool) InitPool(redisPoolCfg *redis.Pool, dial func() (redis.Conn, error)) error {
	if redisPoolCfg == nil {
		if dial == nil {
			return fmt.Errorf("dial is nil")
		}
		redisPoolCfg = &redis.Pool{
			MaxIdle:     10,
			MaxActive:   100,
			IdleTimeout: 240 * time.Second,
			Dial:        dial,
		}
		p.redisPool = redisPoolCfg
		return nil
	}
	if dial != nil {
		if redisPoolCfg.Dial == nil {
			redisPoolCfg.Dial = dial
		}
	}

	p.redisPool = redisPoolCfg
	return nil
}

// ClosePool 关闭连接池
func (p *RedisPool) ClosePool() error {
	if p.redisPool == nil {
		return nil
	}

	err := p.redisPool.Close()
	if err == nil {
		p.redisPool = nil
	}
	return err
}

// GetWithFunc 从连接池中获取一个连接，使用完毕后需要调用conn.Close()关闭连接
func (p *RedisPool) GetWithFunc(ctx context.Context, callFunc func(conn redis.Conn)) error {
	if p.redisPool == nil {
		return fmt.Errorf("pool has closed")
	}

	var tempConn redis.Conn
	if ctx == nil {
		tempConn = p.redisPool.Get()
	} else {
		tempConn2, err := p.redisPool.GetContext(ctx)
		if err != nil {
			return err
		}
		tempConn = tempConn2
	}
	defer func(tempConn redis.Conn) {
		err := tempConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(tempConn)
	callFunc(tempConn)
	return nil
}
