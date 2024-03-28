package microservice

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/94peter/log"
	"github.com/94peter/microservice/cfg"
	"github.com/94peter/microservice/di"
)

type MicroService[T cfg.ModelCfg, R di.ServiceDI] interface {
	GetModelCfgMgr() cfg.ModelCfgMgr
	GetDI() R
	NewLog(name string) (log.Logger, error)
	NewCfg(name string) (T, error)
}

type ServiceHandler func(ctx context.Context)

type microService[T cfg.ModelCfg, R di.ServiceDI] struct {
	Cfg T
	DI  R

	cfgMgr cfg.ModelCfgMgr
}

func New[T cfg.ModelCfg, R di.ServiceDI](mycfg T, mydi R) (MicroService[T, R], error) {
	diCfg, err := di.GetConfigFromEnv()
	if err != nil {
		return nil, err
	}
	err = di.InitServiceDIByCfg(diCfg, mydi)
	if err != nil {
		return nil, err
	}
	if err = mydi.IsConfEmpty(); err != nil {
		return nil, err
	}
	return &microService[T, R]{
		Cfg:    mycfg,
		DI:     mydi,
		cfgMgr: cfg.NewFixModelCfgGinMid(mycfg),
	}, nil
}

func (s *microService[T, R]) GetModelCfgMgr() cfg.ModelCfgMgr {
	return s.cfgMgr
}

func (s *microService[T, R]) GetDI() R {
	return s.DI
}

func (s *microService[T, R]) NewCfg(name string) (T, error) {
	mycfg := s.Cfg.Copy()
	err := mycfg.Init(name, s.DI)
	if err != nil {
		return mycfg.(T), err
	}
	return mycfg.(T), nil
}

func (s *microService[T, R]) NewLog(name string) (log.Logger, error) {
	return s.DI.NewLogger(s.DI.GetService(), name)
}

func RunService(ss ...ServiceHandler) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	var wg sync.WaitGroup
	var ctxCancelSlice = make([]context.CancelFunc, len(ss))
	for i, s := range ss {
		ctx, ctxCancel := context.WithCancel(context.Background())
		ctxCancelSlice[i] = ctxCancel
		wg.Add(1)
		go func(ctx context.Context, handler ServiceHandler) {
			defer wg.Done()
			handler(ctx)
		}(ctx, s)
	}
	<-sig
	for _, cancel := range ctxCancelSlice {
		if cancel != nil {
			cancel()
		}
	}
	wg.Wait()
}
