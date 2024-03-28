package grpc_tool

import (
	"github.com/94peter/microservice/cfg"
	"github.com/94peter/microservice/grpc_tool/interceptor"

	"google.golang.org/grpc"
)

type GrpcConfig struct {
	Port           int  `env:"GRPC_PORT"`
	ReflectService bool `env:"GRPC_REFLECT"`

	Logger              Log
	registerServiceFunc func(grpcServer *grpc.Server)
	interceptors        []interceptor.Interceptor
}

func (c *GrpcConfig) SetRegisterServiceFunc(f func(grpcServer *grpc.Server)) {
	c.registerServiceFunc = f
}

func (c *GrpcConfig) SetInterceptors(i ...interceptor.Interceptor) {
	c.interceptors = i
}

func GetConfigFromEnv() (*GrpcConfig, error) {
	var mycfg GrpcConfig
	err := cfg.GetFromEnv(&mycfg)
	if err != nil {
		return nil, err
	}
	return &mycfg, nil
}

type Log interface {
	Infof(format string, a ...any)
	Fatalf(format string, a ...any)
}
