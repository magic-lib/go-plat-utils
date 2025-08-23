package seqid_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-cache/cache"
	"github.com/magic-lib/go-plat-locker/lock"
	id2 "github.com/magic-lib/go-plat-utils/id-generator/id"
	"github.com/magic-lib/go-plat-utils/id-generator/seqid"
	"testing"
	"time"
)

func TestGeneratorBase32(t *testing.T) {
	aa := id2.GeneratorBase32()
	fmt.Println(aa)
	aa = id2.GetXId()
	fmt.Println(aa)
}
func TestIDGenerator(t *testing.T) {

	ns := "new"

	mysqlCache, err := cache.NewMySQLCache[int64](&cache.MySQLCacheConfig{
		DSN:       "root:tianlin0@tcp(127.0.0.1:3306)/huji?charset=utf8mb4&parseTime=True&loc=Local",
		TableName: "mysql_cache",
		Namespace: ns,
	})
	if err != nil {
		panic(err)
	}

	//mysqlCache := cache.NewMemGoCache[int64](5*time.Minute, 10*time.Minute)

	creator, err := seqid.NewIDGenerator[int64](&seqid.IDGenerator[int64]{
		LockFunc: func(ns, key string) lock.Locker {
			return lock.NewLocker(fmt.Sprintf("%s/%s", ns, key))
		},
		IdStore:   mysqlCache,
		Timeout:   0,
		Namespace: ns,
		BatchSize: 1,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		for i := 0; i < 10; i++ {
			kk, err := creator.Generate("namepb1")
			fmt.Println("test1: ", kk, err)
		}
	}()
	go func() {
		for i := 0; i < 10; i++ {
			kk, err := creator.Generate("namepb1")
			fmt.Println("test2: ", kk, err)
		}
	}()
	go func() {
		for i := 0; i < 15; i++ {
			kk, err := creator.Generate("namepb1")
			fmt.Println("test3: ", kk, err)
		}
	}()

	time.Sleep(5 * time.Second)

}
