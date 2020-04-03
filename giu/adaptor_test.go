// nolint gomnd
package giu_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bingoohuang/gor"

	"github.com/bingoohuang/gor/giu"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.DebugLevel)
}

type Rsp struct {
	State int
	Data  interface{}
}

type SetAgeReq struct {
	Name string
	Age  int
}

type SetAgeRsp struct {
	Name string
}

type AuthUser struct {
	Name string
}

var ga = giu.NewAdaptor()

func init() {
	// 注册如何处理成功返回一个值
	ga.RegisterSuccProcessor(func(c *gin.Context, vs ...interface{}) {
		if len(vs) == 0 {
			c.JSON(http.StatusOK, Rsp{State: http.StatusOK, Data: "ok"}) // 如何处理无返回(单独error返回除外)
		} else if rsp, ok := vs[0].(Rsp); ok { // 返回已经是Rsp类型，不再包装
			c.JSON(http.StatusOK, rsp)
		} else if len(vs) == 1 {
			c.JSON(http.StatusOK, Rsp{State: http.StatusOK, Data: vs[0]}) // 选取第一个返回参数，JSON返回
		} else {
			m := make(map[string]interface{})

			for _, v := range vs {
				m[reflect.TypeOf(v).String()] = v
			}

			c.JSON(http.StatusOK, m)
		}
	})

	// 注册如何处理错误
	ga.RegisterErrProcessor(func(c *gin.Context, vs ...interface{}) {
		c.JSON(http.StatusOK, Rsp{State: http.StatusInternalServerError, Data: vs[0].(error).Error()})
	})

	// 注册如何处理AuthUser类型的输入参数
	ga.RegisterTypeProcessor(AuthUser{}, func(c *gin.Context, vs ...interface{}) (interface{}, error) {
		return gor.V0(c.Get("AuthUser")), nil
	})
}

func TestUMP(t *testing.T) {
	resp := httptest.NewRecorder()
	c, r := gin.CreateTestContext(resp)

	ptrAuthUser := true
	r.Use(func(c *gin.Context) {
		if keep := ga.FindKeep(c); keep != nil {
			logrus.Warnf("keep found %+v", keep)
		}

		if ptrAuthUser {
			c.Set("AuthUser", &AuthUser{Name: "TestAuthUser"})
		} else {
			c.Set("AuthUser", AuthUser{Name: "TestAuthUser"})
		}

		ptrAuthUser = !ptrAuthUser
	})

	type MyObject struct {
		Name string
	}

	gr := ga.Route(r)
	gr.Use(func() *MyObject { return &MyObject{Name: "Test"} })

	gr.GET("/MyObject1", func(m MyObject) string { return m.Name })
	gr.GET("/MyObject2", func(m *MyObject) string { return m.Name })

	gr.GET("/GetAge1/:name", func(user AuthUser, name string) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParams("name")))
	gr.GET("/GetAge2/:name", func(name string, user AuthUser) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParams("name")))
	gr.GET("/GetAge3/:name", func(user *AuthUser, name string) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParams("name")))
	gr.GET("/GetAge4/:name", func(name string, user *AuthUser) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParams("name")))
	gr.POST("/SetAge", func(req SetAgeReq) interface{} {
		return SetAgeRsp{Name: fmt.Sprintf("%s:%d", req.Name, req.Age)}
	})

	gr.Any("/error", func() error { return errors.New("error occurred") })
	gr.GET("/ok", func() error { return nil })
	gr.GET("/url", func(c *gin.Context) string { return c.Request.URL.String() })

	gr.GET("/Get1/:name/:age", f1, giu.Params(giu.URLParams("name", "age")))
	gr.GET("/Get2/:name/:age", f2)
	gr.GET("/Get21/:id", f21)
	gr.GET("/Get3/:name", f3)
	gr.GET("/Get31/:name", f31)
	gr.GET("/Get4", f4)
	gr.GET("/Get51", f51)
	gr.GET("/Get52", f52)
	gr.GET("/Get53/:yes", f53)

	gr.HandleFn(f22)

	ga.RegisterInvokeArounder("logrus", &MyInvokeArounderFactory{})

	gr.HandleFn(f{})

	//r.Run(":8080")

	assertResults(t, resp, c, r)
}

type f struct{}

type f23Url struct {
	giu.T `url:"GET/POST /Get23/:id" logrus:"风起" keep:"ignore"`
}

func (f) Get23(id string, _ f23Url) string { return "hello f23 " + id }

type f24Url struct {
	giu.T `url:"ANY /Get24/:id" logrus:"云涌"`
}

func (f) Get24(id string, _ f24Url) string { return "hello f24 " + id }

// f1 processes /Get1/:name/:age
func f1(name string, age int) (Rsp, error) {
	return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
}

type f22Url struct {
	giu.T `url:"/Get22/:id" logrus:"明月"`
}

func f22(id string, _ f22Url) string { return "hello " + id }

type projectID struct {
	giu.T `arg:"id,url"`
}

// f21 processes /Get21/:id
// gr.GET("/Get21/:id", f21)
func f21(id string, _ projectID) string { return "hello " + id }

// f2 processes /Get2/:name/:age
func f2(name string, age int, _ struct {
	giu.T `arg:"name age,url"` // name和age都是url变量，通过gin.Context.Param(x)获取
}) (Rsp, error) {
	return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
}

// f3 processes  /Get3/:name?age=100
func f3(name string, age int, _ struct {
	_ giu.T `arg:"name,url"`  // name是url变量，通过gin.Context.Param(x)获取
	_ giu.T `arg:"age,query"` // age是query变量，通过gin.Context.Query(x)获取
}) (Rsp, error) {
	return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
}

// f31 processes  /Get3/:name?age=100
func f31(name string, age int, _ struct {
	giu.T `arg:"name,url/age,query"`
}) (Rsp, error) {
	return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
}

// f4 processes /Get4?name=bingoo&age=100
func f4(name string, age int, _ struct {
	giu.T `arg:"name age,query"` // name和age都是query变量，通过gin.Context.Query(x)获取
}) (Rsp, error) {
	return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
}

func f51() giu.DownloadFile {
	return giu.DownloadFile{DiskFile: "testdata/hello.txt"}
}

func f52() interface{} {
	return giu.DownloadFile{Content: []byte("hello"), Filename: "下载.txt"}
}

func f53(yes bool) interface{} {
	if yes {
		return giu.DownloadFile{Content: []byte("hello"), Filename: "下载.txt"}
	}

	return "blabla"
}

func assertResults(t *testing.T, resp *httptest.ResponseRecorder, c *gin.Context, r *gin.Engine) {
	checkStatusOK(t, resp, c, r, "/GetAge1/bingoo", "TestAuthUser/bingoo")
	checkStatusOK(t, resp, c, r, "/GetAge2/bingoo", "TestAuthUser/bingoo")
	checkStatusOK(t, resp, c, r, "/GetAge3/bingoo", "TestAuthUser/bingoo")
	checkStatusOK(t, resp, c, r, "/GetAge4/bingoo", "TestAuthUser/bingoo")
	checkBody(t, resp, c, r, http.MethodPost, "/SetAge", 200,
		SetAgeReq{Name: "bingoo", Age: 100}, SetAgeRsp{Name: "bingoo:100"})
	check(t, resp, c, r, "/error", 500, "error occurred")
	checkStatusOK(t, resp, c, r, "/ok", "ok")
	checkStatusOK(t, resp, c, r, "/Get1/bingoo/100", "bingoo:100")
	checkStatusOK(t, resp, c, r, "/Get2/bingoo/100", "bingoo:100")
	checkStatusOK(t, resp, c, r, "/Get21/bingoo", "hello bingoo")
	checkStatusOK(t, resp, c, r, "/Get22/bingoo", "hello bingoo")
	checkStatusOK(t, resp, c, r, "/Get23/bingoo", "hello f23 bingoo")
	checkStatusOK(t, resp, c, r, "/Get24/bingoo", "hello f24 bingoo")
	checkStatusOK(t, resp, c, r, "/Get3/bingoo?age=100", "bingoo:100")
	checkStatusOK(t, resp, c, r, "/Get31/bingoo?age=100", "bingoo:100")
	checkStatusOK(t, resp, c, r, "/Get4?name=bingoo&age=100", "bingoo:100")
	checkStatusOK(t, resp, c, r, "/url", "/url")
	checkStatusOK(t, resp, c, r, "/MyObject1", "Test")
	checkStatusOK(t, resp, c, r, "/MyObject2", "Test")
	//checkStatusOK(t, resp, c, r, "/Get5", "Test")

	c.Request, _ = http.NewRequest(http.MethodGet, "/Get51", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "bingoohuang", strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/Get52", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello", strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/Get53/true", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, `hello`, strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/Get53/false", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"State":200,"Data":"blabla"}`, strings.TrimSpace(string(body)))
}

func checkStatusOK(t *testing.T, rr *httptest.ResponseRecorder, c *gin.Context, r *gin.Engine, url string, d interface{}) {
	check(t, rr, c, r, url, http.StatusOK, d)
}

func check(t *testing.T, rr *httptest.ResponseRecorder, c *gin.Context, r *gin.Engine, url string, state int, d interface{}) {
	checkBody(t, rr, c, r, http.MethodGet, url, state, nil, d)
}
func checkBody(t *testing.T, rr *httptest.ResponseRecorder, c *gin.Context, r *gin.Engine,
	method, url string, state int, b interface{}, d interface{}) {
	if b != nil {
		bb, _ := json.Marshal(b)
		c.Request, _ = http.NewRequest(method, url, bytes.NewReader(bb))
		c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")
	} else {
		c.Request, _ = http.NewRequest(method, url, nil)
	}

	r.ServeHTTP(rr, c.Request)
	assert.Equal(t, http.StatusOK, rr.Code)
	rsp, _ := json.Marshal(Rsp{State: state, Data: d})
	body, _ := ioutil.ReadAll(rr.Body)

	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))
}

func TestHello(t *testing.T) {
	router := gin.New()
	hello := ""
	world := ""

	ga := giu.NewAdaptor()
	gr := ga.Route(router)

	gr.GET("/hello/:arg", func(v string) { hello = v }, giu.Params(giu.URLParams("arg")))
	gr.GET("/world", func(v string) string {
		world = v
		return "hello " + v
	}, giu.Params(giu.QueryParams("arg")))

	gr.GET("/error", func() error {
		return errors.New("xxx")
	})

	gr.GET("/Get54", func() interface{} { return giu.DirectResponse{Code: 203} })
	gr.GET("/Get55", func() interface{} { return &giu.DirectResponse{Code: 203} })

	rr := performRequest("GET", "/hello/bingoo", router)

	assert.Equal(t, "bingoo", hello)
	assert.Equal(t, gor.V([]byte{}, nil), gor.V(ioutil.ReadAll(rr.Body)))

	rr = performRequest("GET", "/world?arg=huang", router)

	assert.Equal(t, "huang", world)
	content, _ := ioutil.ReadAll(rr.Body)
	assert.Equal(t, `hello huang`, strings.TrimSpace(string(content)))

	rr = performRequest("GET", "/error", router)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	rr = performRequest("GET", "/Get54", router)
	assert.Equal(t, 203, rr.Code)

	rr = performRequest("GET", "/Get55", router)
	assert.Equal(t, 203, rr.Code)
}

// from https://github.com/gin-gonic/gin/issues/1120
func performRequest(method, target string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

// InvokeArounderFactory defines the factory to create InvokeArounder
type MyInvokeArounderFactory struct {
}

// Create creates the InvokeArounder with the tag value and the handler type.
func (MyInvokeArounderFactory) Create(handlerName, tag string, handler giu.HandlerFunc) giu.InvokeArounder {
	return &myInvokeArounder{handlerName: handlerName, Tag: tag, Handler: handler}
}

// myInvokeArounder defines the adaptee invoking before and after intercepting points for the user.
type myInvokeArounder struct {
	Tag         string
	Handler     giu.HandlerFunc
	handlerName string
}

// Before will be called before the adaptee invoking.
func (a *myInvokeArounder) Before(args []interface{}) error {
	logrus.Debugf("invoke %s before %s with args %v", a.handlerName, a.Tag, args)
	return nil
}

// After will be called after the adaptee invoking.
func (a *myInvokeArounder) After(outs []interface{}) {
	logrus.Debugf("invoke %s after %s with outs %v", a.handlerName, a.Tag, outs)
}
