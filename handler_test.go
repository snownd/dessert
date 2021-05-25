package dessert

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testPathParam struct {
	IPathParam
	Foo string `uri:"foo"`
}

type testDTO struct {
	IDTO
	Foo string `json:"foo" form:"foo" binding:"required"`
	Bar string `json:"bar" form:"bar" `
}

type testHeader struct {
	IHeader
	Foo string `header:"Header-Foo"`
}

type testResponse struct {
	*JsonBaseResponse
	Foo string
	Bar string
}

func TestInjectPathParamFn(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	foo := "bar"
	c.Params = append(c.Params, gin.Param{Key: "foo", Value: foo})
	tp := reflect.TypeOf(testPathParam{})
	v, err := injectPathParamFn(tp, c)
	if assert.NoError(t, err) {
		param, ok := v.Interface().(*testPathParam)
		assert.True(t, ok)
		assert.Equal(t, foo, param.Foo)
	}
}

func TestInjectDTOFnWithJSON(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"foo\":\"bar\", \"bar\":\"foo\"}"))
	c.Request.Header.Add("Content-Type", gin.MIMEJSON)
	tp := reflect.TypeOf(testDTO{})
	v, err := injectDTOFn(tp, c)
	if assert.NoError(t, err) {
		dto, ok := v.Interface().(*testDTO)
		assert.True(t, ok)
		assert.Equal(t, "foo", dto.Bar)
		assert.Equal(t, "bar", dto.Foo)
	}
}

func TestInjectDTOFnWithForm(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("foo=bar&bar=foo"))
	c.Request.Header.Add("Content-Type", gin.MIMEPOSTForm)
	tp := reflect.TypeOf(testDTO{})
	v, err := injectDTOFn(tp, c)
	if assert.NoError(t, err) {
		dto, ok := v.Interface().(*testDTO)
		assert.True(t, ok)
		assert.Equal(t, "foo", dto.Bar)
		assert.Equal(t, "bar", dto.Foo)
	}
}

func TestINjectDTOfnWithQuery(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?foo=bar&bar=foo", nil)
	tp := reflect.TypeOf(testDTO{})
	v, err := injectDTOFn(tp, c)
	if assert.NoError(t, err) {
		dto, ok := v.Interface().(*testDTO)
		assert.True(t, ok)
		assert.Equal(t, "foo", dto.Bar)
		assert.Equal(t, "bar", dto.Foo)
	}
}

func TestINjectDTOfnWithBindError(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?bar=foo", nil)
	tp := reflect.TypeOf(testDTO{})
	_, err := injectDTOFn(tp, c)
	assert.ErrorIs(t, err, ErrBind)
}

func TestInjectHeaderFn(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("foo=bar&bar=foo"))
	c.Request.Header.Add("Header-Foo", "bar")
	tp := reflect.TypeOf(testHeader{})
	v, err := injectHeaderFn(tp, c)
	if assert.NoError(t, err) {
		h, ok := v.Interface().(*testHeader)
		assert.True(t, ok)
		assert.Equal(t, "bar", h.Foo)
	}
}

func TestNewHandlerFuncWithEmptyParam(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	h, err := newHandler(MethodGet, "/", func() {}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "/", h.Path)
		assert.Equal(t, MethodGet, h.Method)
		h.Fn(c)
		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	}
}

func TestNewHanderFuncWithParams(t *testing.T) {
	resHandler = func(ctx *Context, status int, res IResponse) {
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, gin.MIMEJSON, res.ContentType())
		data, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"Foo":"bar","Bar":"foo"}`, string(data))
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/?foo=bar&bar=foo", nil)
	c.Request.Header.Add("Header-Foo", "bar")
	h, err := newHandler(MethodGet, "/", func(header *testHeader, dto *testDTO) (int, IResponse, error) {
		assert.Equal(t, "bar", header.Foo)
		assert.Equal(t, "foo", dto.Bar)
		assert.Equal(t, "bar", dto.Foo)
		return http.StatusOK, &testResponse{Res, dto.Foo, dto.Bar}, nil
	}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "/", h.Path)
		assert.Equal(t, MethodGet, h.Method)
		h.Fn(c)
	}
}
