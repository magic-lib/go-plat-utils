package commerror_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/commerror"
	"google.golang.org/grpc/codes"
	"testing"
)

func TestError(t *testing.T) {
	err := commerror.GrpcError(codes.OK, "aaaaa")
	t.Log(err)
	err1, code := commerror.FromGrpcError(err)
	t.Log(err1, code)

	err = commerror.GrpcError(51005, "aaaaa")
	t.Log(err)
	err1, code = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(code, code.String())

	err = commerror.GrpcError(codes.NotFound, "aaaaa", commerror.New("bbbbb", 51005))
	t.Log(err)
	err1, code = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(code, code.String())

	err = commerror.New("aaaaaaaffffff", 10051)
	err1, code = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(code, code.String())

	err = fmt.Errorf("aaaaa")
	err1, code = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(code, code.String())

}
