package dessert

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	Fn      RequestHandlerFunc
	Path    string
	Method  RequestMethod
	Options RequestHandlerOptions
}

type RequestHandlerOptions struct {
	HandlerInterceptor []HandlerInterceptor
}

// injectValueFromGINCtxFn return value representing a pointer of type t
type injectValueFromGINCtxFn = func(t reflect.Type, ctx *gin.Context) (reflect.Value, error)
type handlerArgType struct {
	t         reflect.Type
	isPointer bool
	fn        injectValueFromGINCtxFn
}

var param *IPathParam
var paramType = reflect.TypeOf(param).Elem()
var dto *IDTO
var dtoType = reflect.TypeOf(dto).Elem()
var resp *IResponse
var respType = reflect.TypeOf(resp).Elem()
var header *IHeader
var headerType = reflect.TypeOf(header).Elem()
var e *error
var errType = reflect.TypeOf(e).Elem()
var contextPtr *context.Context
var contextType = reflect.TypeOf(contextPtr).Elem()

var injectPathParamFn injectValueFromGINCtxFn = func(t reflect.Type, ctx *gin.Context) (reflect.Value, error) {
	vPtr := reflect.New(t)
	err := ctx.ShouldBindUri(vPtr.Interface())
	if err != nil {
		return vPtr, fmt.Errorf("%w for Path: %s, detail: %v", ErrBind, ctx.FullPath(), err)
	}
	return vPtr, nil
}

var injectDTOFn injectValueFromGINCtxFn = func(t reflect.Type, ctx *gin.Context) (reflect.Value, error) {
	vPtr := reflect.New(t)
	if ctx.ContentType() == gin.MIMEJSON {
		err := ctx.ShouldBindJSON(vPtr.Interface())
		if err != nil {
			return vPtr, fmt.Errorf("%w for DTO: %v", ErrBind, err.Error())
		}
		return vPtr, nil
	}
	err := ctx.ShouldBind(vPtr.Interface())
	if err != nil {
		return vPtr, fmt.Errorf("%w for DTO: %s", ErrBind, err.Error())
	}
	return vPtr, nil
}

var injectHeaderFn injectValueFromGINCtxFn = func(t reflect.Type, ctx *gin.Context) (reflect.Value, error) {
	vPtr := reflect.New(t)
	err := ctx.ShouldBindHeader(vPtr.Interface())
	if err != nil {
		return vPtr, fmt.Errorf("%w for Header: %s", ErrBind, err.Error())
	}
	return vPtr, nil
}

func newHandler(method RequestMethod, path string, fn interface{}, opts *RequestHandlerOptions) (*RequestHandler, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("dessert: param fn of handler(method=%s,path=%s) must be a function", method, path)
	}
	if fnType.IsVariadic() {
		return nil, fmt.Errorf("dessert: param fn of handler(method=%s,path=%s) must be a variadic function", method, path)
	}
	inputTypes, withContext, err := getInputTypes(fnValue, fnType)
	if err != nil {
		return nil, fmt.Errorf("%w , handler: method=%s, path=%s", err, method, path)
	}
	if err = checkOutput(fnType); err != nil {
		return nil, fmt.Errorf("%w , handler: method=%s, path=%s", err, method, path)
	}
	rh := &RequestHandler{
		Method: method,
		Path:   path,
		Fn:     newHandlerFunc(fnValue, fnType, inputTypes, withContext),
	}
	if opts != nil {
		rh.Options = *opts
	}

	return rh, nil
}

func getInputTypes(fnValue reflect.Value, fnType reflect.Type) ([]*handlerArgType, bool, error) {
	numIn := fnType.NumIn()
	inputs := make([]*handlerArgType, numIn)
	withContext := false
	for i := 0; i < numIn; i++ {
		t := fnType.In(i)
		isPointer := false
		if t.Kind() == reflect.Ptr {
			isPointer = true
			t = t.Elem()
		}
		var injectFn injectValueFromGINCtxFn

		if t.Implements(contextType) {
			if i != 0 {
				return nil, withContext, errors.New("dessert:context must be first parameter")
			}
			withContext = true
			continue
		} else if t.Implements(dtoType) {
			injectFn = injectDTOFn
		} else if t.Implements(paramType) {
			injectFn = injectPathParamFn
		} else if t.Implements(headerType) {
			injectFn = injectHeaderFn
		} else {
			return nil, withContext, errors.New("dessert:unknown param type")
		}
		inputs[i] = &handlerArgType{t, isPointer, injectFn}

	}
	return inputs, withContext, nil
}

func checkOutput(fnType reflect.Type) error {
	numOut := fnType.NumOut()
	if numOut != 3 && numOut != 0 {
		return errors.New("dessert: handler fn must have 3 or 0 output parameters")
	}
	if numOut == 3 {
		if fnType.Out(0).Kind() != reflect.Int {
			return errors.New("dessert: handler fn first output must be an integer")
		}
		if !fnType.Out(1).Implements(respType) {
			return errors.New("dessert: handler fn second output must implement IResponse")
		}
		if !fnType.Out(2).Implements(errType) {
			return errors.New("dessert: handler fn third output must implement error")
		}
	}
	return nil
}

func newHandlerFunc(fnValue reflect.Value, fnType reflect.Type, inputTypes []*handlerArgType, withContext bool) RequestHandlerFunc {
	return func(ctx *Context) {
		numIn := fnType.NumIn()
		numOut := fnType.NumOut()
		args := make([]reflect.Value, numIn)
		i := 0
		if withContext {
			i = 1
			args[0] = reflect.ValueOf(ctx.Request.Context())
		}
		for ; i < numIn; i++ {
			argType := inputTypes[i]
			v, err := argType.fn(argType.t, ctx)
			if err != nil {
				errHandler(ctx, err)
				return
			}
			if argType.isPointer {
				args[i] = v
			} else {
				args[i] = v.Elem()
			}
		}
		out := fnValue.Call(args)
		if numOut == 3 {
			if !out[2].IsNil() {
				errHandler(ctx, (out[2].Elem().Interface()).(error))
				return
			}
			status := int(out[0].Int())
			resHandler(ctx, status, (out[1].Interface()).(IResponse))
		} else {
			ctx.Data(http.StatusNoContent, gin.MIMEJSON, nil)
		}
	}
}
