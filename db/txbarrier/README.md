# About txbarrier

txbarrier 用于解决分布式事务中由于网络延迟或错误等，导致分支请求 RM(Resource Manager) 可能出现的时序问题：
- 请求幂等：分布式事务协调器（TC）可能由于请求 RM 时发生超时成功（从 TC 角度是失败，从 RM 角度是成功），导致 TC 重试对 RM 的请求，RM 需要保证请求的幂等性；
- 事务空回滚：事务分支执行时可能由于网络延迟出现 Cancel 比 Try 请求先到达 RM 的情况，RM 需要正确处理这种 Cancel 请求；
- 事务悬挂：当 Cancel 比 Try 先到达 RM 之后，Try 再到达 RM，RM 需要识别并处理这种异常。

# Usage

## SQL

以 MySQL 为例子，底层驱动实现使用 [github.com/go-sql-driver/mysql](github.com/go-sql-driver/mysql)。
首先建立 txbarrier 所需要的记录表：
```sql
create table if not exists tdxa.txbarrier (
    `id` bigint(22) PRIMARY KEY AUTO_INCREMENT,
    `xid` varchar(128) default '',
    `branch_id` varchar(128) default '',
    `op` varchar(45) default '',
    `reason` varchar(45) default '',
    UNIQUE KEY `barrier_unique_key`(`xid`, `branch_id`, `op`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
```
使用代码如下：
```go
package main

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
	"github.com/magic-lib/go-plat-utils/db/txbarrier/sqlbarrier"
)

func main() {
	// 基于 go-sql-driver/mysql 注册 sqlbarrier 驱动
	driverName := "sqlbarrier-mysql"
	sql.Register(driverName, sqlbarrier.NewDriver(&mysql.MySQLDriver{}))

	db, err := sql.Open(driverName, "username:password@tcp(host:port)/Dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}

	// 将信息注入到请求 context 中
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      "123",
		BranchID: "example",
		TransTyp: "tcc",
		Op:       txbarrier.Try, // Try/Confirm/Cancel
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		if errors.Is(err, txbarrier.ErrDuplicationOrSuspension) {
			// 重复请求或者发生了悬挂的情况
			return
        }
		if errors.Is(err, txbarrier.ErrEmptyCompensation) {
			// 发生空回滚
			return
		}
		return
    }

	// 执行业务逻辑
	ret, err := tx.ExecContext(ctx, "update test_table set content = ? where identify=?", "content", "001")
	if err != nil {
		tx.Rollback()
		return
	}

	// ...

	tx.Commit()
}
```

## MongoDB

首先建立 txbarrier 所需要的 collection：
```
use tdxa
db.txbarrier.drop()
db.txbarrier.createIndex({ xid: 1, branch_id: 1, op: 1 }, { unique: true })
db.txbarrier.insert({ xid: "123", branch_id: "01", op: "try", reason: "cancel" })
```
使用代码如下：
```go
package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
	"github.com/magic-lib/go-plat-utils/db/txbarrier/mgobarrier"
)

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return
	}

	// 将信息注入到请求 context 中
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      "123",
		BranchID: "example",
		TransTyp: "tcc",
		Op:       txbarrier.Try, // Try/Confirm/Cancel
	})

	err = mgobarrier.DoWithClient(ctx, client, func(sc mongo.SessionContext) error {
		// db 操作，必须使用 sc
		ret, err := client.Database("db").Collection("collection").InsertOne(sc, bson.D{
			{Key: "content", Value: "content"},
			{Key: "identify", Value: "test"},
		})
		if err != nil {
			return err
		}

		// 其他业务逻辑

		return nil
	})

	if err != nil {
		if errors.Is(err, txbarrier.ErrDuplicationOrSuspension) {
			// 重复请求或者发生了悬挂的情况
			return
		}
		if errors.Is(err, txbarrier.ErrEmptyCompensation) {
			// 发生空回滚
			return
		}
		return
	}
}
```

## Redis

txbarrier 的实现原理主要是依靠数据库的本地事务，通过在一个数据库事务中建立屏障数据和执行业务数据操作，达到识别处理分布式事务中可能存在的分支调用时序问题。 而在Redis中，通常意义上的事务是指以`MULTI`, `EXEC`, `DISCARD` 和 `WATCH`为核心的，让Redis批量执行一系列命令的操作。在实际应用中，使用Redis Lua 脚本实现Redis事务，是业务方更加普遍的选择：Lua脚本比 `MULTI` 等命令更加简单高效。
因此 txbarrier 优先选择支持 Redis Lua 脚本的事务屏障，目前仅支持 eval、evalsha 命令，欢迎有兴趣的开发一起完善共建。

**Note:** Redis事务/Lua脚本本身不支持失败数据回滚，在一个 Lua 脚本中，如果业务数据操作了一半才失败需要回滚的，应该由业务的 Lua 脚本负责回滚数据，txbarrier 无法帮助业务回滚业务数据。

使用代码如下：
```go
package main

import (
	"context"
	"fmt"
	
	"github.com/redis/go-redis/v9"
	
	"github.com/magic-lib/go-plat-utils/db/txbarrier"
	"github.com/magic-lib/go-plat-utils/db/txbarrier/rdsbarrier"
)

// testScript 演示逻辑，运行程序前，需要保证 KEYS[1] 的 value 是数值， 否则lua脚本会执行失败: set testKey 0
// 业务的 Lua 脚本必须有两个返回值，当第二个返回值为 "SUCCESS" 时，屏障才会认为业务逻辑成功执行，否则认为业务逻辑失败
// 注意：演示脚本中没有失败需要回滚业务数据的需要，真实业务中，如果脚本是执行到一半才发现需要终止，那么已经写的数据应该在脚本
//      中由业务负责回滚。
const testScript = `
local v = redis.call('GET', KEYS[1])

if v == false or v + ARGV[1] < 0 then
	return {'', 'FAILURE'}
end

local ret = redis.call('INCRBY', KEYS[1], ARGV[1])
return {ret, 'SUCCESS'}
`

func main() {
	cli := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cli.AddHook(rdsbarrier.NewHook(rdsbarrier.WithTimeout(3600)))

	// 如果使用 evalsha 命令执行脚本，需要提前加载
	// scriptSha, _ := cli.ScriptLoad(context.TODO(), rdsbarrier.WrapScript(testScript)).Result()
	
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      "abc",
		BranchID: "123",
		TransTyp: "tcc",
		Op:       txbarrier.Try,
	})
	
	// 可以使用 EvalSha 或 Eval （注意：EvalSha 需要提前载入脚本）
	// result, err := cli.EvalSha(ctx, scriptSha, []string{"balance"}, 1).Result()
	result, err := cli.Eval(ctx, testScript, []string{"balance"}, 1).Result()
	if err != nil {
		if errors.Is(err, txbarrier.ErrDuplicationOrSuspension) {
			// 重复请求或者发生了悬挂的情况
			return
		}
		if errors.Is(err, txbarrier.ErrEmptyCompensation) {
			// 发生空回滚
			return
		}
		return
	}
	
	fmt.Println(result)
	// Output:
	// [1 SUCCESS]
}

```

由于 `rdsbarrier.Hook` 要实现屏障逻辑，需要对 redis 请求进行一些脚本插入和命令解析，对性能有一定损耗，简单的 Benchmark 对比如下：
```shell
$ go test -bench='Hook$' -benchmem .
goos: darwin
goarch: amd64
pkg: github.com/magic-lib/go-plat-utils/db/txbarrier/rdsbarrier
cpu: VirtualApple @ 2.50GHz
BenchmarkHook-10           	    3693	    329795 ns/op	  441558 B/op	    1652 allocs/op
BenchmarkWithoutHook-10    	    4658	    220224 ns/op	  386576 B/op	    1483 allocs/op
PASS
ok  	github.com/magic-lib/go-plat-utils/db/txbarrier/rdsbarrier	3.702s
```
**注意：** 此处 Benchmark 仅供业务参考，实际性能表现需要业务根据业务逻辑进行压力测试对比。