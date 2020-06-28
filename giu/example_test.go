package giu_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bingoohuang/gor/giu"
	"github.com/bingoohuang/gou/lo"
	"github.com/gin-gonic/gin"
)

type Name struct {
	Name string `json:"name"`
}

// nolint:gochecknoinits
func init() {
	lo.SetupLog()
}

func ExampleNewAdaptor() {
	router := gin.Default()
	gr := giu.NewAdaptor().Route(router)

	gr.GET("/hello/:arg", func(v string) string {
		return "Welcome " + v
	})

	rr := PerformRequest("GET", "/hello/bingoo", router)
	fmt.Println(rr.Body.String())

	// Output:
	// Welcome bingoo
}

func ExampleAdaptorError() {
	router := gin.Default()
	gr := giu.NewAdaptor().Route(router)

	gr.GET("/error", func(v string) error {
		return errors.New("error occurred")
	})

	rr := PerformRequest("GET", "/error", router)
	fmt.Println(rr.Code)
	// Output:
	// 500
}

func ExampleErr() {
	router := gin.Default()
	ga := giu.NewAdaptor()
	gr := ga.Route(router)

	// 注册如何处理错误
	ga.RegisterErrProcessor(func(c *gin.Context, vs ...interface{}) {
		err := vs[0].(error)
		giu.Jsonify(c, http.StatusOK, Rsp{State: http.StatusInternalServerError, Data: err.Error()})
	})

	gr.GET("/error", func(v string) error {
		return errors.New("error occurred")
	})

	rr := PerformRequest("GET", "/error", router)
	fmt.Println(rr.Body.String())
	// Output:
	// {"state":500,"data":"error occurred"}
}

func ExampleJSON() {
	router := gin.Default()
	gr := giu.NewAdaptor().Route(router)

	gr.POST("/JSON0", func(j Name) string { return "JSON " + j.Name })
	gr.POST("/JSON2", func(j Name) Name { j.Name = "Hello " + j.Name; return j })
	gr.POST("/JSON21", func(j Name) (Name, error) { j.Name = "Hello " + j.Name; return j, nil })
	gr.POST("/JSON22", func(j Name) (Name, error) { return Name{}, errors.New("rejected " + j.Name) })

	r3 := PerformRequest("POST", "/JSON0", router, JSONObject(Name{Name: "bingoohuang"}))
	fmt.Println(r3.Body.String())

	r4 := PerformRequest("POST", "/JSON2", router, JSONObject(Name{Name: "bingoohuang"}))
	fmt.Println(r4.Body.String())

	r5 := PerformRequest("POST", "/JSON21", router, JSONObject(Name{Name: "bingoohuang"}))
	fmt.Println(r5.Body.String())

	r6 := PerformRequest("POST", "/JSON22", router, JSONObject(Name{Name: "bingoohuang"}))
	fmt.Println(r6.Code)

	// Output:
	// JSON bingoohuang
	// {"name":"Hello bingoohuang"}
	//
	// {"name":"Hello bingoohuang"}
	//
	// 500
}
