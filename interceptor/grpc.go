package interceptor

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"runtime/debug"
	"time"
)

type GrpcServer struct {
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

func (_ *GrpcServer) LoggingUnaryServerInterceptor(
	logBegin func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo),
	logAfter func(ctx context.Context, startTime time.Time, resp interface{}, err error)) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		startTime := time.Now()
		if logBegin != nil {
			logBegin(ctx, req, info)
		}
		//md, _ := metadata.FromIncomingContext(ctx)
		resp, err = handler(ctx, req)
		if logAfter != nil {
			logAfter(ctx, startTime, resp, err)
		}
		return resp, err
	}
}

func (_ *GrpcServer) TracerServerOption() grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler())
}
func (_ *GrpcServer) RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Errorf(codes.Internal, "panic in method=%s, err=%v, stack=%s",
					info.FullMethod, r, debug.Stack())
			}
		}()
		return handler(ctx, req)
	}
}

func (_ *GrpcServer) AuthUnaryServerInterceptor(authCheck func(ctx context.Context, md metadata.MD) (context.Context, error)) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if authCheck == nil {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		newCtx, err := authCheck(ctx, md)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return handler(newCtx, req)
	}
}

func (_ *GrpcServer) LoggingStreamServerInterceptor(
	wrappedStream func(ss grpc.ServerStream) grpc.ServerStream, logBegin func(ss grpc.ServerStream, info *grpc.StreamServerInfo),
	logAfter func(startTime time.Time, err error)) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()
		ws := ss
		if wrappedStream != nil {
			ws = wrappedStream(ss)
		}
		if logBegin != nil {
			logBegin(ws, info)
		}
		err := handler(srv, ws)

		if logAfter != nil {
			logAfter(startTime, err)
		}
		return err
	}
}

type GrpcClient struct {
}

func NewGrpcClient() *GrpcClient {
	return &GrpcClient{}
}

func (_ *GrpcClient) TraceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		traceID := ctx.Value("traceId")
		if traceID == nil {
			traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
			ctx = context.WithValue(ctx, "traceId", traceID)
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		md.Set("x-trace-id", traceID.(string))
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (_ *GrpcClient) RetryUnaryClientInterceptor(maxRetry int, wait time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		for i := 0; i <= maxRetry; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			st, ok := status.FromError(err)
			if !ok {
				return err
			}

			// 只对部分错误重试
			if st.Code() != codes.Unavailable && st.Code() != codes.DeadlineExceeded {
				return err
			}

			// 等一会再重试，简单点就固定睡
			time.Sleep(wait)
		}
		return err
	}
}
