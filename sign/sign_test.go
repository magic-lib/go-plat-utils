package sign

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGeneratorSignature(t *testing.T) {
	oneMap := map[string]any{
		//"a":   1,
		//"b":   "2",
		//"c":   `/ ? # & = @ : , ; + - * () * ~!@#$%^&*()_+,./';:"?><`,
		//"d":   true,
		//"aaa": "wo 是谁",
	}
	signService, _ := New(&BaseSignCheck{
		GetAppTokenFunc: func(ctx context.Context, appId string) (string, error) {
			return "secret", nil
		},
		GetNonceCreateTimeFunc: func(ctx context.Context, appId, nonce string) (time.Time, error) {
			return time.Now(), nil
		},
	})
	newParam, err := signService.GeneratorSignature(context.Background(), &BaseSignDto{
		//ClientId:     "admin",
		//ClientSecret: "secret",
		Timestamp:  "1234567890",
		Nonce:      "1234567890",
		SignMethod: "sha-256",
	}, oneMap)
	fmt.Println(newParam, err)
}
