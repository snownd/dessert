package dessert

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	GroupController       = "controller"
	GroupGlobalMiddleware = "serverMiddleware"
)

var ErrBind = errors.New("Dessert:BindError")

type RequestMethod string

const (
	MethodGet     RequestMethod = http.MethodGet
	MethodPost    RequestMethod = http.MethodPost
	MethodPut     RequestMethod = http.MethodPut
	MethodDelete  RequestMethod = http.MethodDelete
	MethodHead    RequestMethod = http.MethodHead
	MethodOptions RequestMethod = http.MethodOptions
	MethodPatch   RequestMethod = http.MethodPatch
	MethodTrace   RequestMethod = http.MethodTrace
)

type Context = gin.Context

type RequestHandlerFunc func(*Context)

type Middleware = gin.HandlerFunc

type HandlerInterceptor = Middleware

// IPathParam http path param
// tag: name:"id"
type IPathParam interface {
	path() string
}

// IDTO http query or body
type IDTO interface {
	raw() []byte
}

// IHeader http request header
type IHeader interface {
	header() map[string]string
}

// IResponse http response
type IResponse interface {
	ContentType() string
	Reader() io.ReadCloser
	ContentLength() int64
}

type JsonBaseResponse struct {
}

func (r JsonBaseResponse) ContentType() string {
	return gin.MIMEJSON
}

func (r JsonBaseResponse) Reader() io.ReadCloser {
	return nil
}

func (r JsonBaseResponse) ContentLength() int64 {
	return 0
}

var Res = &JsonBaseResponse{}
