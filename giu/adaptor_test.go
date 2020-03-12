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
	"strings"
	"testing"

	"github.com/bingoohuang/goreflect"

	"github.com/bingoohuang/goreflect/giu"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestUMP(t *testing.T) {
	ga := giu.NewAdaptor()

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

	// 注册如何处理成功返回一个值
	ga.RegisterTypeProcessor(giu.SuccessfullyInvokedType, func(c *gin.Context, vs ...interface{}) (interface{}, error) {
		if len(vs) == 0 {
			c.JSON(http.StatusOK, Rsp{State: http.StatusOK, Data: "ok"}) // 如何处理无返回(单独error返回除外)
		} else if rsp, ok := vs[0].(Rsp); ok { // 返回已经是Rsp类型，不再包装
			c.JSON(http.StatusOK, rsp)
		} else {
			c.JSON(http.StatusOK, Rsp{State: http.StatusOK, Data: vs[0]}) // 选取第一个返回参数，JSON返回
		}

		return nil, nil
	})

	// 注册如何处理错误
	ga.RegisterTypeProcessor(goreflect.ErrType, func(c *gin.Context, vs ...interface{}) (interface{}, error) {
		c.JSON(http.StatusOK, Rsp{State: http.StatusInternalServerError, Data: vs[0].(error).Error()})

		return nil, nil
	})

	// 注册如何处理AuthUser类型的输入参数
	ga.RegisterTypeProcessor(AuthUser{}, func(c *gin.Context, vs ...interface{}) (interface{}, error) {
		authUser, _ := c.Get("AuthUser")

		return authUser, nil
	})

	resp := httptest.NewRecorder()
	c, r := gin.CreateTestContext(resp)

	r.Use(func(c *gin.Context) {
		c.Set("AuthUser", &AuthUser{Name: "TestAuthUser"})
	})

	r.GET("/GetAge1/:name", ga.Adapt(func(user AuthUser, name string) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParam("name"))))
	r.GET("/GetAge2/:name", ga.Adapt(func(name string, user AuthUser) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParam("name"))))
	r.GET("/GetAge3/:name", ga.Adapt(func(user *AuthUser, name string) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParam("name"))))
	r.GET("/GetAge4/:name", ga.Adapt(func(name string, user *AuthUser) string {
		return user.Name + "/" + name
	}, giu.Params(giu.URLParam("name"))))
	r.POST("/SetAge", ga.Adapt(func(req SetAgeReq) SetAgeRsp {
		return SetAgeRsp{Name: fmt.Sprintf("%s:%d", req.Name, req.Age)}
	}))

	r.GET("/Get/:name/:age", ga.Adapt(func(name string, age int) (Rsp, error) {
		return Rsp{State: 200, Data: fmt.Sprintf("%s:%d", name, age)}, nil
	}, giu.Params(giu.URLParam("name"), giu.URLParam("age"))))

	r.GET("/error", ga.Adapt(func() error { return errors.New("error occurred") }))

	r.GET("/ok", ga.Adapt(func() error { return nil }))

	c.Request, _ = http.NewRequest(http.MethodGet, "/GetAge1/bingoo", nil)
	r.ServeHTTP(resp, c.Request)

	rsp, _ := json.Marshal(Rsp{State: http.StatusOK, Data: "TestAuthUser/bingoo"})
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/GetAge2/bingoo", nil)
	r.ServeHTTP(resp, c.Request)

	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: "TestAuthUser/bingoo"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/GetAge3/bingoo", nil)
	r.ServeHTTP(resp, c.Request)

	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: "TestAuthUser/bingoo"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/GetAge4/bingoo", nil)
	r.ServeHTTP(resp, c.Request)

	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: "TestAuthUser/bingoo"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	req, _ := json.Marshal(SetAgeReq{Name: "bingoo", Age: 100})
	c.Request, _ = http.NewRequest(http.MethodPost, "/SetAge", bytes.NewReader(req))
	c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")
	r.ServeHTTP(resp, c.Request)

	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: SetAgeRsp{Name: "bingoo:100"}})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/error", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	rsp, _ = json.Marshal(Rsp{State: http.StatusInternalServerError, Data: "error occurred"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: "ok"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/Get/bingoo/100", nil)
	r.ServeHTTP(resp, c.Request)
	assert.Equal(t, http.StatusOK, resp.Code)
	rsp, _ = json.Marshal(Rsp{State: http.StatusOK, Data: "bingoo:100"})
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(rsp), strings.TrimSpace(string(body)))
}

func TestHello(t *testing.T) {
	ga := giu.NewAdaptor()
	resp := httptest.NewRecorder()
	c, r := gin.CreateTestContext(resp)
	hello := ""
	world := ""

	r.GET("/hello/:arg", ga.Adapt(func(v string) { hello = v }, giu.Params(giu.URLParam("arg"))))
	r.GET("/world", ga.Adapt(func(v string) { world = v }, giu.Params(giu.QueryParam("arg"))))

	c.Request, _ = http.NewRequest(http.MethodGet, "/hello/bingoo", nil)
	r.ServeHTTP(resp, c.Request)

	assert.Equal(t, "bingoo", hello)
	assert.Equal(t, m2([]byte{}, nil), m2(ioutil.ReadAll(resp.Body)))

	c.Request, _ = http.NewRequest(http.MethodGet, "/world", nil)
	q := c.Request.URL.Query()
	q.Add("arg", "huang")
	c.Request.URL.RawQuery = q.Encode()

	r.ServeHTTP(resp, c.Request)

	assert.Equal(t, "huang", world)
	assert.Equal(t, m2([]byte{}, nil), m2(ioutil.ReadAll(resp.Body)))
}

func m2(v ...interface{}) []interface{} {
	return v
}
