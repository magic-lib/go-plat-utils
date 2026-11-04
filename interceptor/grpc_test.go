package interceptor_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func TestGrpcServer(t *testing.T) {
	if os.Getenv("GO_PLAT_INTEGRATION") == "" {
		t.Skip("集成测试，监听 50051 端口持续提供服务；设置 GO_PLAT_INTEGRATION=1 后运行")
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen err: %v", err)
	}

	grpcServer := interceptor.NewGrpcServer()

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcServer.RecoveryUnaryServerInterceptor(),
			grpcServer.AuthUnaryServerInterceptor(nil),
			grpcServer.LoggingUnaryServerInterceptor(nil, nil),
		),
		grpc.ChainStreamInterceptor(
			grpcServer.LoggingStreamServerInterceptor(nil, nil, nil),
		),
		grpcServer.TracerServerOption(),
	)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcServer.LoggingUnaryServerInterceptor(nil, nil)),
	)

	fmt.Println(server)

	log.Println("grpc server started on :50051")

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve err: %v", err)
	}
}
func TestGrpcClient(t *testing.T) {
	if os.Getenv("GO_PLAT_INTEGRATION") == "" {
		t.Skip("集成测试，需要 127.0.0.1:50051 的 gRPC 服务；设置 GO_PLAT_INTEGRATION=1 后运行")
	}
	grpcClient := interceptor.NewGrpcClient()

	conn, err := grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcClient.TraceUnaryClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	conn, err = grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcClient.TraceUnaryClientInterceptor(),
			grpcClient.RetryUnaryClientInterceptor(2, 200*time.Millisecond),
		),
	)

	fmt.Println(conn, err)
}
