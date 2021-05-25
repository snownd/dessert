package dessert

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type GlobalErrorHandler = func(ctx *Context, err error)

type GlobalResponseHandler = func(ctx *Context, status int, res IResponse)

type GlobalMiddleware Middleware

type ServerNetOptions struct {
	Timeout        time.Duration
	Addr           string
	MaxHeaderBytes int
}

type ServerOptions struct {
	fx.In
	Controllers     []Controller       `group:"controller"`
	Middlewares     []GlobalMiddleware `group:"serverMiddleware"`
	ErrorHandler    GlobalErrorHandler
	ResponseHandler GlobalResponseHandler
	NetOptions      ServerNetOptions
}

var defaultGlobalErrorHandler GlobalErrorHandler = func(ctx *Context, err error) {
	if errors.Is(err, ErrBind) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"message": "bindError",
			"detail":  err.Error(),
		})
	} else {
		ctx.Abort()
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

var defaultGlobalReponseHandler GlobalResponseHandler = func(ctx *Context, status int, res IResponse) {
	if res == nil {
		ctx.Data(http.StatusNoContent, gin.MIMEJSON, nil)
		return
	}
	switch res.ContentType() {
	case gin.MIMEJSON:
		ctx.JSON(status, res)
	case gin.MIMEXML, gin.MIMEXML2:
		ctx.XML(status, res)
	default:
		ctx.DataFromReader(status, res.ContentLength(), res.ContentType(), res.Reader(), nil)
	}
}

func NewDefaultErrorHandler() GlobalErrorHandler {
	return defaultGlobalErrorHandler
}

func NewDefaultResponseHandler() GlobalResponseHandler {
	return defaultGlobalReponseHandler
}

func NewServerNetOptions() ServerNetOptions {
	return ServerNetOptions{
		Timeout:        30 * time.Second,
		Addr:           ":3000",
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
}

func NewRecoveryMiddleware() GlobalMiddleware {
	return GlobalMiddleware(gin.Recovery())
}

func NewGlobalMiddlewareContainer(constructor interface{}) fx.Annotated {
	return fx.Annotated{
		Group:  GroupGlobalMiddleware,
		Target: constructor,
	}
}

var DefaultServerOptions = fx.Options(
	fx.Provide(NewDefaultErrorHandler),
	fx.Provide(NewDefaultResponseHandler),
	fx.Provide(NewGlobalMiddlewareContainer(NewRecoveryMiddleware)),
	fx.Provide(NewServerNetOptions),
)
