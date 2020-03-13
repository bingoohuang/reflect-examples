package giu

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/bingoohuang/gor"
	"github.com/gin-gonic/gin"
)

// TypeProcessor is the processor for a specified type.
type TypeProcessor func(*gin.Context, ...interface{}) (interface{}, error)

// Processor is the processor for a specified type.
type Processor func(*gin.Context, ...interface{})

// Adaptor is the adaptor structure for gin.HandlerFunc.
type Adaptor struct {
	register map[reflect.Type]TypeProcessor
}

// NewAdaptor makes a new Adaptor.
func NewAdaptor() *Adaptor {
	return &Adaptor{
		register: make(map[reflect.Type]TypeProcessor),
	}
}

// RegisterErrProcessor register a type processor for the error.
func (a *Adaptor) RegisterErrProcessor(p Processor) {
	a.register[gor.ErrType] = func(c *gin.Context, args ...interface{}) (interface{}, error) {
		p(c, args...)
		return nil, nil
	}
}

// RegisterSuccProcessor register a type processor for the successful deal.
func (a *Adaptor) RegisterSuccProcessor(p Processor) {
	a.register[GetNonPtrType(SuccInvokedType)] = func(c *gin.Context, args ...interface{}) (interface{}, error) {
		p(c, args...)
		return nil, nil
	}
}

// RegisterTypeProcessor register a type processor for the type.
func (a *Adaptor) RegisterTypeProcessor(t interface{}, p TypeProcessor) {
	a.register[GetNonPtrType(t)] = p
}

func (a *Adaptor) findTypeProcessor(t reflect.Type) TypeProcessor {
	for k, v := range a.register {
		if gor.ImplType(t, k) {
			return v
		}
	}

	return nil
}

func (a *Adaptor) findTypeProcessorOr(t reflect.Type, defaultProcessor TypeProcessor) TypeProcessor {
	p := a.findTypeProcessor(t)
	if p != nil {
		return p
	}

	return defaultProcessor
}

// SuccInvoked represents the adaptor is invoked successfully with returned value.
type SuccInvoked interface {
	SuccessfullyInvoked()
}

var (
	// SuccInvokedType defines the successfully adaptor invoke's type.
	// nolint gochecknoglobals
	SuccInvokedType = reflect.TypeOf((*SuccInvoked)(nil)).Elem()
)

// GetNonPtrType returns the non-ptr type of v.
func GetNonPtrType(v interface{}) reflect.Type {
	if vt, ok := v.(reflect.Type); ok {
		return vt
	}

	var t reflect.Type

	if vt, ok := v.(reflect.Value); ok {
		t = vt.Type()
	} else {
		t = reflect.TypeOf(v)
	}

	if t.Kind() != reflect.Ptr {
		return t
	}

	return t.Elem()
}

// StateCodeError represents the error with a StateCode.
type StateCodeError interface {
	GetStateCode() int
}

func defaultSuccessProcessor(g *gin.Context, vs ...interface{}) (interface{}, error) {
	if len(vs) > 0 {
		v0 := vs[0]
		if reflect.Indirect(reflect.ValueOf(v0)).Kind() == reflect.Struct {
			g.JSON(http.StatusOK, v0)
		} else {
			g.String(http.StatusOK, fmt.Sprintf("%v", v0))
		}
	}

	return nil, nil
}

func defaultErrorProcessor(g *gin.Context, vs ...interface{}) (interface{}, error) {
	code := 500
	if sce, ok := vs[0].(StateCodeError); ok {
		code = sce.GetStateCode()
	}

	_ = g.AbortWithError(code, vs[0].(error))

	return nil, nil
}

// Param defines the interface to how to get param's value.
type Param interface {
	Get(g *gin.Context) string
}

type urlParam struct {
	key string
}

// URLParam defines the URL param in the URL PATH.
func URLParam(key string) Param {
	return urlParam{key: key}
}

func (u urlParam) Get(g *gin.Context) string {
	return g.Param(u.key)
}

var _ Param = (*urlParam)(nil)

type queryParam struct {
	key          string
	defaultValue string
}

// QueryParam defines the query param.
func QueryParam(key string) Param {
	return QueryParamOr(key, "")
}

// QueryParamOr defines the query param and the default value when empty.
func QueryParamOr(key, defaultValue string) Param {
	return queryParam{key: key, defaultValue: defaultValue}
}

func (u queryParam) Get(g *gin.Context) string {
	return g.DefaultQuery(u.key, u.defaultValue)
}

var _ Param = (*queryParam)(nil)

// Option defines the adatpor's option.
type Option struct {
	Params     []Param
	MiddleWare bool
}

// OptionFn is the function prototype to apply option.
type OptionFn func(*Option)

// Params defines the params for the adaptor.
func Params(ps ...Param) OptionFn { return func(option *Option) { option.Params = ps } }

// MiddleWare defines the middleWare flag for the adaptor.
func MiddleWare(m bool) OptionFn { return func(option *Option) { option.MiddleWare = m } }

// Adapt adapts convenient function to gi.HandleFunc.
func (a *Adaptor) Adapt(fn HandlerFunc, optionFns ...OptionFn) gin.HandlerFunc {
	option := &Option{}

	for _, f := range optionFns {
		f(option)
	}

	fv := reflect.ValueOf(fn)
	errTp := a.findTypeProcessorOr(gor.ErrType, defaultErrorProcessor)

	return func(c *gin.Context) {
		argVs, err := a.createArgs(c, fv, option)
		if err != nil {
			_, _ = errTp(c, err)

			return
		}

		r := fv.Call(argVs)

		if err := a.processOut(c, fv, r, option); err != nil {
			_, _ = errTp(c, err)
		}
	}
}

func (a *Adaptor) processOut(c *gin.Context, fv reflect.Value, r []reflect.Value, option *Option) error {
	ft := fv.Type()
	numOut := ft.NumOut()

	if numOut == 0 {
		return nil
	}

	if gor.AsError(ft.Out(numOut - 1)) { // nolint gomnd
		if !r[numOut-1].IsNil() {
			return r[numOut-1].Interface().(error)
		}

		numOut-- // drop the error returned by the adapted.
	}

	vs := make([]interface{}, numOut)

	a.registerInjects(c, option, numOut, r)

	for i := 0; i < numOut; i++ {
		vs[i] = r[i].Interface()
	}

	if option.MiddleWare {
		return nil
	}

	p := a.findTypeProcessorOr(SuccInvokedType, defaultSuccessProcessor)
	_, _ = p(c, vs...)

	return nil
}

func (a *Adaptor) registerInjects(c *gin.Context, option *Option, numOut int, r []reflect.Value) {
	if !option.MiddleWare {
		return
	}

	for i := 0; i < numOut; i++ {
		nonPtrRi := convertPtr(false, r[i])
		if nonPtrRi.Kind() == reflect.Struct {
			c.Set("_inject_"+nonPtrRi.Type().String(), convertPtr(true, r[i]))
		}
	}
}

func (a *Adaptor) createArgs(c *gin.Context, fv reflect.Value, option *Option) ([]reflect.Value, error) {
	ft := fv.Type()
	numIn := ft.NumIn()

	ii := -1
	argVs := make([]reflect.Value, numIn)
	argTags := a.findTags(ft)

	for i := 0; i < numIn; i++ {
		argType, argKind, isArgTypePtr := parseArgs(ft, i)

		switch argKind {
		case reflect.Struct:
			if _, isTagArg := argTags[i]; isTagArg {
				argVs[i] = reflect.Zero(_TType)
			} else {
				v, err := a.processStruct(c, argType, isArgTypePtr)
				if err != nil {
					return nil, err
				}

				argVs[i] = convertPtr(isArgTypePtr, v)
			}
		default:
			ii++
			v, err := dealDirectParamArg(c, option.Params[ii], argKind)

			if err != nil {
				return nil, err
			}

			argVs[i] = convertPtr(isArgTypePtr, reflect.ValueOf(v))
		}
	}

	return argVs, nil
}

func (a *Adaptor) findTags(ft reflect.Type) map[int][]reflect.StructTag {
	argTags := make(map[int][]reflect.StructTag)

	for i := 0; i < ft.NumIn(); i++ {
		argType, argKind, _ := parseArgs(ft, i)
		if argKind != reflect.Struct {
			continue
		}

		if tags := findTags(argType, _TType); len(tags) > 0 {
			argTags[i] = tags
		}
	}

	return argTags
}

func findTags(t reflect.Type, target reflect.Type) []reflect.StructTag {
	tags := make([]reflect.StructTag, 0)
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if tf.Type == target {
			tags = append(tags, tf.Tag)
		}
	}

	return tags
}

func dealDirectParamArg(c *gin.Context, param Param, argKind reflect.Kind) (interface{}, error) {
	argValue := param.Get(c)

	switch argKind {
	case reflect.String:
		return argValue, nil
	case reflect.Int:
		return strconv.Atoi(argValue)
	}

	return nil, fmt.Errorf("unsupported type %v", argKind)
}

func parseArgs(ft reflect.Type, argIndex int) (reflect.Type, reflect.Kind, bool) {
	argType := ft.In(argIndex)
	isArgTypePtr := argType.Kind() == reflect.Ptr

	if isArgTypePtr {
		argType = argType.Elem()
	}

	return argType, argType.Kind(), isArgTypePtr
}

func (a *Adaptor) processStruct(c *gin.Context, argType reflect.Type, isArgTypePtr bool) (reflect.Value, error) {
	if isArgTypePtr && argType == GetNonPtrType(c) { // 直接注入gin.Context
		return reflect.ValueOf(c), nil
	}

	if v, exists := c.Get("_inject_" + argType.String()); exists {
		return convertPtr(isArgTypePtr, v.(reflect.Value)), nil
	}

	if tp := a.findTypeProcessor(argType); tp != nil {
		v, err := tp(c, argType)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(v), nil
	}

	argValue := reflect.New(argType)
	if err := c.ShouldBind(argValue.Interface()); err != nil {
		return reflect.Value{}, err
	}

	return argValue, nil
}

func convertPtr(isPtr bool, v reflect.Value) reflect.Value {
	if !isPtr {
		return reflect.Indirect(v)
	}

	if v.Kind() == reflect.Ptr {
		return v
	}

	p := reflect.New(v.Type())
	p.Elem().Set(v)

	return p
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware
// that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (a *Routes) Handle(httpMethod, relativePath string, h HandlerFunc, fns ...OptionFn) IRoutes {
	a.GinRoutes.Handle(httpMethod, relativePath, a.Adaptor.Adapt(h, fns...))
	return a
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (a *Routes) POST(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("POST", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (a *Routes) GET(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("GET", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (a *Routes) DELETE(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("DELETE", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (a *Routes) PATCH(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("PATCH", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (a *Routes) PUT(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("PUT", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (a *Routes) OPTIONS(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("OPTIONS", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (a *Routes) HEAD(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Handle("HEAD", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (a *Routes) Any(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRoutes.Any(relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc interface {
}

// IRoutes defines all router handle interface.
type IRoutes interface {
	Use(HandlerFunc, ...OptionFn) IRoutes

	Handle(string, string, HandlerFunc, ...OptionFn) IRoutes
	Any(string, HandlerFunc, ...OptionFn) IRoutes
	GET(string, HandlerFunc, ...OptionFn) IRoutes
	POST(string, HandlerFunc, ...OptionFn) IRoutes
	DELETE(string, HandlerFunc, ...OptionFn) IRoutes
	PATCH(string, HandlerFunc, ...OptionFn) IRoutes
	PUT(string, HandlerFunc, ...OptionFn) IRoutes
	OPTIONS(string, HandlerFunc, ...OptionFn) IRoutes
	HEAD(string, HandlerFunc, ...OptionFn) IRoutes
}

// Route makes a route for Adaptor.
func (a *Adaptor) Route(r gin.IRoutes) IRoutes {
	return &Routes{GinRoutes: r, Adaptor: a}
}

// Routes defines adaptor routes implemetation for IRoutes.
type Routes struct {
	GinRoutes gin.IRoutes
	Adaptor   *Adaptor
}

// Use adds middleware, see example code.
func (a *Routes) Use(h HandlerFunc, optionFns ...OptionFn) IRoutes {
	fns := make([]OptionFn, len(optionFns)+1) // nolint gomnd
	copy(fns, optionFns)
	fns[len(optionFns)] = MiddleWare(true)
	a.GinRoutes.Use(a.Adaptor.Adapt(h, fns...))

	return a
}

var _ IRoutes = (*Routes)(nil)

// T defines the tag for handler functions.
type T interface {
	HandlerFnTag()
}

var (
	// _TType defines the type of T.
	// nolint gochecknoglobals
	_TType = reflect.TypeOf((*T)(nil)).Elem()
)
