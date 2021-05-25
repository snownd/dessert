package dessert

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Server struct {
	hs *http.Server
}

var errHandler GlobalErrorHandler
var resHandler GlobalResponseHandler

func NewServer(lc fx.Lifecycle, opts ServerOptions) (*Server, error) {
	errHandler = opts.ErrorHandler
	resHandler = opts.ResponseHandler
	engine := gin.New()
	for _, itc := range opts.Middlewares {
		engine.Use(gin.HandlerFunc(itc))
	}
	for _, ctr := range opts.Controllers {
		group := engine.Group(ctr.Path)
		if len(ctr.Middlewares) > 0 {
			group.Use(ctr.Middlewares...)
		}
		for _, h := range ctr.Handlers {
			group.Handle(string(h.Method), h.Path, gin.HandlerFunc(h.Fn))
		}
	}
	hs := &http.Server{
		Addr:           opts.NetOptions.Addr,
		Handler:        engine,
		MaxHeaderBytes: opts.NetOptions.MaxHeaderBytes,
		ReadTimeout:    opts.NetOptions.Timeout,
	}
	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			fmt.Printf("[dessert] start HTTP server on addr:%s \n", hs.Addr)
			go hs.ListenAndServe()
			return nil
		},
		OnStop: func(c context.Context) error {
			fmt.Println("[dessert] stopping HTTP server...")
			return hs.Shutdown(c)
		},
	})
	return &Server{hs}, nil
}

var DefaultServerModule = fx.Options(
	DefaultServerOptions,
	fx.Provide(NewServer),
	fx.Invoke(func(s *Server) {}),
)
