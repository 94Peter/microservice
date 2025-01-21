package mid

import (
	"github.com/94peter/microservice/apitool/err"
	"github.com/gin-gonic/gin"
)

type GinMiddle interface {
	err.ErrorHandler
	Handler() gin.HandlerFunc
}

func NewGinMiddle(handler gin.HandlerFunc) GinMiddle {
	return &baseMiddle{
		handler: handler,
	}
}

type baseMiddle struct {
	handler gin.HandlerFunc
	err.CommonErrorHandler
}

func (m *baseMiddle) Handler() gin.HandlerFunc {
	return m.handler
}
