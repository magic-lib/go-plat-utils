package goroutines_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"testing"
	"time"
)

func ExampleGoAsync() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "fish", "章鱼")

	goroutines.SetContext(&ctx)

	goroutines.OpenRoutinePool(9)
	defer goroutines.CloseRoutinePool()

	_, pId, _ := goroutines.GetContext()
	fmt.Println("start: " + pId)

	for i := 0; i < 10; i++ {
		goroutines.GoAsync(func(params ...interface{}) {
			fmt.Printf("i=%d \n", params[0].(int))
			c, pId, _ := goroutines.GetContext()
			fmt.Println(c, pId)
			if c != nil {
				fmt.Println((*c).Value("fish"), pId)
			}
			c1 := context.WithValue(*c, "pig", "猪")
			goroutines.SetContext(&c1)
		}, i)
	}

	time.Sleep(2 * time.Second)

	c2, pId2, _ := goroutines.GetContext()
	fmt.Println("pig2:", (*c2).Value("pig"), (*c2).Value("fish"), pId2)

	goroutines.GoAsync(func(params ...interface{}) {
		c3, pId3, _ := goroutines.GetContext()
		if c3 != nil {
			fmt.Println("pig3:", (*c3).Value("pig"), (*c3).Value("fish"), pId3)
		}

	})

	time.Sleep(1 * time.Second)

	goroutines.DelContext()

	// Output:
}

func TestExampleGoSync(t *testing.T) {
	start := time.Now()
	_, _ = goroutines.GoAsyncTimeout(3*time.Second, func(params ...interface{}) (bool, error) {
		time.Sleep(2 * time.Second)
		fmt.Println("test")
		return true, nil

	})

	fmt.Println(time.Since(start).Seconds())

	time.Sleep(3 * time.Second)

}
