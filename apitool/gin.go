package apitool

import (
	"github.com/94peter/microservice/apitool/err"
	"github.com/gin-gonic/gin"
)

type GinHandler struct {
	Handler func(c *gin.Context)
	Method  string
	Path    string
}

type GinAPI interface {
	err.ErrorHandler
	GetHandlers() []*GinHandler
}
