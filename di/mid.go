package di

import (
	"context"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func GinMiddleHandler(di DI) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetDiToGin(c, di)
		c.Next()
	}
}

func GrpcStreamInterceptor(di DI) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if isReflectMethod(info.FullMethod) {
			return handler(srv, ss)
		}
		ctx := SetDiToCtx(ss.Context(), di)

		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func GrpcUnaryInterceptor(di DI) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = SetDiToCtx(ctx, di)
		return handler(ctx, req)
	}
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func isReflectMethod(m string) bool {
	return m == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"
}
