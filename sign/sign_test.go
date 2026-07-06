package sign

import (
	"context"
	"fmt"
	cryptotool "github.com/magic-lib/go-plat-utils/crypto"
	"net/http"
	"testing"
	"time"
)

func TestGeneratorSignature(t *testing.T) {
	//oneMap := map[string]any{
	//	//"a":   1,
	//	//"b":   "2",
	//	//"c":   `/ ? # & = @ : , ; + - * () * ~!@#$%^&*()_+,./';:"?><`,
	//	//"d":   true,
	//	//"aaa": "wo 是谁",
	//}
	signService, _ := New(&BaseSignCheck{
		GetAppTokenFunc: func(ctx context.Context, appId string) (string, error) {
			return getAppToken(appId), nil
		},
		GetOrSetNonceFunc: func(ctx context.Context, appId, nonce string, timeout time.Duration) (int, error) {
			return 0, nil
		},
	})

	newParam, err := signService.GeneratorSignature(context.Background(), &BaseSignDto{
		AppId:      "biu-tech",
		Timestamp:  "1783090590",
		Nonce:      "AoFgOQ",
		SignMethod: "MD5",
	}, nil)
	fmt.Println(newParam, err)

	//app_token:ebbbb80d535e6c06cf64707ada1bdfd1
	//signature:a819ac39486203969569d0595fe96aa9
}

func getAppToken(appId string) string {
	newToken := fmt.Sprintf("ussd/%s", appId)
	name := cryptotool.Md5(newToken)
	return cryptotool.Md5(name)
}

func TestCheckSignature(t *testing.T) {
	baseSignCheckService, err := New(&BaseSignCheck{
		HeaderKeyPrefix: "X-",
		GetAppTokenFunc: func(ctx context.Context, appId string) (string, error) {
			return getAppToken(appId), nil
		},
		NoCheckTimestamp: true,
	})
	if err != nil {
		return
	}
	checked, err := baseSignCheckService.CheckSignature(context.Background(), http.Header{
		"X-App-Id":      []string{"biu-tech"},
		"X-Timestamp":   []string{"1783090590"},
		"X-Nonce":       []string{"AoFgOQ"},
		"X-Signature":   []string{"a819ac39486203969569d0595fe96aa9"},
		"X-Sign-Method": []string{"MD5"},
	}, nil)
	if err != nil {
		return
	}
	fmt.Println(checked)

	//app_token:ebbbb80d535e6c06cf64707ada1bdfd1

	//app_id=biu%2Dtech&app_token=ebbbb80d535e6c06cf64707ada1bdfd1&nonce=AoFgOQ&sign_method=MD5&timestamp=1783090590

	oldStr := fmt.Sprintf("app_id=biu-tech&app_token=ebbbb80d535e6c06cf64707ada1bdfd1&nonce=AoFgOQ&sign_method=MD5&timestamp=1783090590")

	//app_id=biu%2Dtech&app_token=ebbbb80d535e6c06cf64707ada1bdfd1&nonce=AoFgOQ&sign_method=MD5&timestamp=1783090590
	signToken := cryptotool.Md5(oldStr)
	fmt.Println(signToken)

}
