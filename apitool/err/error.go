package err

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GinServiceErrorHandler func(c *gin.Context, service string, err error)

type GinErrorHandler func(c *gin.Context, err error)

type GinErrorWithStatusHandler func(c *gin.Context, status int, err error)

type ApiError interface {
	GetStatus() int
	error
}

type myApiError struct {
	statusCode int
	error
}

func (e myApiError) GetStatus() int {
	return e.statusCode
}

func (e myApiError) String() string {
	return fmt.Sprintf("%v: %v", e.statusCode, e.error)
}

func New(status int, msg string) ApiError {
	return myApiError{statusCode: status, error: errors.New(msg)}
}

func PkgError(status int, err error) ApiError {
	return myApiError{statusCode: status, error: err}
}

type ErrorHandler interface {
	SetErrorHandler(GinErrorHandler)
}

type CommonErrorHandler struct {
	GinErrorHandler
}

func (api *CommonErrorHandler) SetErrorHandler(errHandler GinErrorHandler) {
	api.GinErrorHandler = errHandler
}

func (api *CommonErrorHandler) GinErrorWithStatusHandler(c *gin.Context, status int, err error) {
	api.GinErrorHandler(c, PkgError(status, err))
}
