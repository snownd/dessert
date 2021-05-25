package dessert

import "go.uber.org/fx"

type Controller struct {
	Path        string
	Middlewares []Middleware
	Handlers    []RequestHandler
}

func NewControllerContainer(constructor interface{}) fx.Annotated {
	return fx.Annotated{
		Group:  GroupController,
		Target: constructor,
	}
}

func NewController(path string, middlewares []Middleware) Controller {
	return Controller{
		Path:        path,
		Middlewares: middlewares,
		Handlers:    make([]RequestHandler, 0),
	}
}

func NewBaseController(path string) Controller {
	return NewController(path, nil)
}

func (c *Controller) GetWithOpt(path string, fn interface{}, opts *RequestHandlerOptions) *Controller {
	c.newHandler(MethodGet, path, fn, opts)
	return c
}

func (c *Controller) Get(path string, fn interface{}) *Controller {
	return c.GetWithOpt(path, fn, nil)
}

func (c *Controller) PostWithOpt(path string, fn interface{}, opts *RequestHandlerOptions) *Controller {
	c.newHandler(MethodPost, path, fn, opts)
	return c
}

func (c *Controller) Post(path string, fn interface{}) *Controller {
	return c.PostWithOpt(path, fn, nil)
}

func (c *Controller) PutWithOpt(path string, fn interface{}, opts *RequestHandlerOptions) *Controller {
	c.newHandler(MethodPut, path, fn, opts)
	return c
}

func (c *Controller) Put(path string, fn interface{}) *Controller {
	return c.PutWithOpt(path, fn, nil)
}

func (c *Controller) DeleteWithOpt(path string, fn interface{}, opts *RequestHandlerOptions) *Controller {
	c.newHandler(MethodDelete, path, fn, opts)
	return c
}

func (c *Controller) Delete(path string, fn interface{}) *Controller {
	return c.DeleteWithOpt(path, fn, nil)
}

func (c *Controller) Options(path string, fn interface{}, opts *RequestHandlerOptions) {
	c.newHandler(MethodOptions, path, fn, opts)
}

func (c *Controller) Head(path string, fn interface{}, opts *RequestHandlerOptions) {
	c.newHandler(MethodHead, path, fn, opts)
}

func (c *Controller) Patch(path string, fn interface{}, opts *RequestHandlerOptions) {
	c.newHandler(MethodPatch, path, fn, opts)
}

func (c *Controller) Trace(path string, fn interface{}, opts *RequestHandlerOptions) {
	c.newHandler(MethodTrace, path, fn, opts)
}

func (c *Controller) newHandler(method RequestMethod, path string, fn interface{}, opts *RequestHandlerOptions) {
	h, err := newHandler(method, path, fn, opts)
	if err != nil {
		panic(err)
	}
	c.Handlers = append(c.Handlers, *h)
}
