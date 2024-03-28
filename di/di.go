package di

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/94peter/log"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v3"
)

type ctxKey string

const (
	_KEY_DI = "SERVICE_DI"

	_CTX_DI = ctxKey(_KEY_DI)
)

type DI interface {
	IsConfEmpty() error
	GetService() string
}

type ServiceDI interface {
	DI
	setService(string)
	log.LoggerDI
}

type CommonServiceDI struct {
	service string
}

func (s *CommonServiceDI) GetService() string {
	return s.service
}

func (impl *CommonServiceDI) setService(s string) {
	impl.service = s
}

func SetDiToCtx[T DI](ctx context.Context, di T) context.Context {
	return context.WithValue(ctx, _CTX_DI, di)
}

func GetDiFromCtx[T DI](ctx context.Context) T {
	var di T
	val := ctx.Value(_CTX_DI)
	if data, ok := val.(T); ok {
		return data
	}
	return di
}

func SetDiToGin[T DI](c *gin.Context, di T) {
	c.Set(_KEY_DI, di)
}

func GetDiFromGin[T DI](c *gin.Context) T {
	var di T
	val, ok := c.Get(_KEY_DI)
	if !ok {
		return di
	}
	if data, ok := val.(T); ok {
		return data
	}
	return di
}

func InitConfByFile(f string, di DI) error {
	yamlFile, err := os.ReadFile(f)
	if err != nil {
		return errors.New("load conf fail: " + f)
	}
	return InitConfByByte(yamlFile, di)
}

func InitConfByByte(b []byte, di DI) error {
	err := yaml.Unmarshal(b, di)
	if err != nil {
		return err
	}
	return nil
}

const confFileTpl = "%s%s/%s"

// 初始化設定檔，讀YAML檔
func IniConfByEnv(path, env, fname string, di DI) error {
	return InitConfByFile(fmt.Sprintf(confFileTpl, path, env, fname), di)
}

func InitConf(path, fname string, di DI) error {
	return InitConfByFile(path+fname, di)
}

func InitConfByCfg(cfg *config, di DI) error {
	return InitConfByFile(cfg.File, di)
}

func InitServiceDIByCfg(cfg *config, di ServiceDI) error {
	err := InitConfByFile(cfg.File, di)
	if err != nil {
		return err
	}
	name, err := os.Hostname()
	if err != nil {
		di.setService(fmt.Sprintf("%s-%s", cfg.Service, name))
	} else {
		di.setService(cfg.Service)
	}
	return nil
}

func InitConfByUri(uri string, di DI) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read conf fail")
	}
	err = yaml.Unmarshal(body, di)
	if err != nil {
		return err
	}
	return nil
}

type config struct {
	Service string
	File    string
}

func GetConfigFromEnv() (*config, error) {
	cfg := config{
		Service: os.Getenv("SERVICE"),
		File:    os.Getenv("CONFIG_FILE"),
	}
	if cfg.Service == "" {
		return nil, errors.New("SERVICE is empty")
	}
	if cfg.File == "" {
		return nil, errors.New("CONFIG_FILE is empty")
	}
	return &cfg, nil
}
