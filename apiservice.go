package microservice

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/94peter/microservice/apitool"
	"github.com/94peter/microservice/apitool/err"
	"github.com/94peter/microservice/apitool/mid"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type ginServ struct {
	*gin.Engine
	service    string
	errHandler err.GinServiceErrorHandler
	mids       []mid.GinMiddle
	apis       []apitool.GinAPI
	debug      bool
}

func (g *ginServ) defaultErrorHandler(c *gin.Context, service string, myerr error) {
	apiErr, ok := myerr.(err.ApiError)
	resp := gin.H{"service": service, "error": myerr.Error()}
	if ok {
		c.JSON(apiErr.GetStatus(), resp)
	} else {
		c.JSON(http.StatusInternalServerError, resp)
	}
}

func (g *ginServ) errorHandler(c *gin.Context, err error) {
	g.errHandler(c, g.service, err)
}

func (g *ginServ) init() {
	for _, m := range g.mids {
		m.SetErrorHandler(g.errorHandler)
	}
	router := g.Use(g.getMiddles()...)

	for _, a := range g.apis {
		a.SetErrorHandler(g.errorHandler)
		for _, h := range a.GetHandlers() {
			router.Handle(h.Method, h.Path, h.Handler)
		}
	}
}

func (g *ginServ) getMiddles() []gin.HandlerFunc {
	var middles []gin.HandlerFunc
	if g.debug {
		middles = append(middles, mid.DebugHandler())
	}
	for _, m := range g.mids {
		middles = append(middles, m.Handler())
	}
	return middles

}

type options func(*ginServ)

func WithErrorHandler(errHandler err.GinServiceErrorHandler) options {
	return func(g *ginServ) {
		g.errHandler = errHandler
	}
}

func WithMiddle(mids ...mid.GinMiddle) options {
	return func(g *ginServ) {
		g.mids = mids
	}
}

func WithAPI(apis ...apitool.GinAPI) options {
	return func(g *ginServ) {
		g.apis = apis
	}
}

func WithPromhttp(c ...prometheus.Collector) options {
	return func(g *ginServ) {
		prometheus.MustRegister(c...)
		g.GET("/metrics", func(c *gin.Context) { promhttp.Handler().ServeHTTP(c.Writer, c.Request) }).Use()
	}
}

func NewApiWithViper(opts ...options) (ServiceHandler, error) {
	service := viper.GetString("service")
	if service == "" {
		return nil, errors.New("service is empty")
	}
	port := viper.GetUint("api.port")
	if port == 0 {
		return nil, errors.New("api.port is empty")
	}
	var mode string
	debug := viper.GetBool("api.debug")
	if debug {
		mode = "debug"
	} else {
		mode = "release"
	}

	gin.SetMode(mode)
	serv := &ginServ{
		Engine:  gin.New(),
		service: service,
		debug:   debug,
	}
	for _, opt := range opts {
		opt(serv)
	}

	if serv.errHandler == nil {
		serv.errHandler = serv.defaultErrorHandler
	}

	serv.init()

	return func(ctx context.Context) {
		log.Println("start api service port:", port)
		runApiService(ctx, &http.Server{
			Addr:    ":" + strconv.Itoa(int(port)),
			Handler: serv.Engine,
		})
	}, nil
}

func runApiService(ctx context.Context, serv *http.Server) {
	var apiWait sync.WaitGroup
	const fiveSecods = 5 * time.Second
	apiWait.Add(1)
	go func(srv *http.Server) {

		for {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				time.Sleep(fiveSecods)
			} else if err == http.ErrServerClosed {
				apiWait.Done()
				return
			}
		}
	}(serv)

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), fiveSecods)
	defer cancel()
	if err := serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	apiWait.Wait()
}
