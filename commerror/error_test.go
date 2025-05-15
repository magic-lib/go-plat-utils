package commerror_test

import (
	"github.com/magic-lib/go-plat-utils/commerror"
	"google.golang.org/grpc/codes"
	"testing"
)

func TestError(t *testing.T) {
	err := commerror.GrpcError(codes.OK, "aaaaa")
	t.Log(err)
	err1, err2 := commerror.FromGrpcError(err)
	t.Log(err1, err2)

	err = commerror.GrpcError(51005, "aaaaa")
	t.Log(err)
	err1, err2 = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(err2, err2.String())

	err = commerror.GrpcError(codes.NotFound, "aaaaa", commerror.New("bbbbb", 51005))
	t.Log(err)
	err1, err2 = commerror.FromGrpcError(err)
	t.Log(err1.Error(), err1.Code())
	t.Log(err2, err2.String())

}
