package giu

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/bingoohuang/goreflect"
	"github.com/gin-gonic/gin"
)

// TypeProcessor is the processor for a specified type.
type TypeProcessor func(*gin.Context, ...interface{}) (interface{}, error)

// Adaptor is the adaptor structure for gin.HandlerFunc.
type Adaptor struct {
	register map[reflect.Type]TypeProcessor
}

// NewAdaptor makes a new Adaptor.
func NewAdaptor() *Adaptor {
	return &Adaptor{register: make(map[reflect.Type]TypeProcessor)}
}

// RegisterTypeProcessor register a type processor for the type.
func (a *Adaptor) RegisterTypeProcessor(t interface{}, p TypeProcessor) {
	a.register[GetNonPtrType(t)] = p
}

func (a *Adaptor) findTypeProcessor(t reflect.Type) TypeProcessor {
	for k, v := range a.register {
		if goreflect.ImplType(t, k) {
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

// SuccessfullyInvoked represents the adaptor is invoked successfully with returned value.
type SuccessfullyInvoked interface {
	SuccessfullyInvoked()
}

var (
	// SuccessfullyInvokedType defines the successfully adaptor invoke's type.
	// nolint gochecknoglobals
	SuccessfullyInvokedType = reflect.TypeOf((*SuccessfullyInvoked)(nil)).Elem()
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
		g.JSON(http.StatusOK, vs[0])
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
	Params []Param
}

// OptionFn is the function prototype to apply option.
type OptionFn func(*Option)

// Params defines the params for the adaptor.
func Params(params ...Param) OptionFn {
	return func(option *Option) {
		option.Params = params
	}
}

// Adapt adapts convenient function to gi.HandleFunc.
func (a *Adaptor) Adapt(fn interface{}, optionFns ...OptionFn) gin.HandlerFunc {
	option := &Option{}

	for _, f := range optionFns {
		f(option)
	}

	fv := reflect.ValueOf(fn)
	errTp := a.findTypeProcessorOr(goreflect.ErrType, defaultErrorProcessor)

	return func(c *gin.Context) {
		argVs, err := a.createArgs(c, fv, option)
		if err != nil {
			_, _ = errTp(c, err)

			return
		}

		r := fv.Call(argVs)

		if err := a.processOut(c, fv, r); err != nil {
			_, _ = errTp(c, err)
		}
	}
}

func (a *Adaptor) processOut(c *gin.Context, fv reflect.Value, r []reflect.Value) error {
	ft := fv.Type()
	numOut := ft.NumOut()

	if numOut == 0 {
		return nil
	}

	if goreflect.AsError(ft.Out(numOut - 1)) { // nolint gomnd
		if !r[numOut-1].IsNil() {
			return r[numOut-1].Interface().(error)
		}

		numOut-- // drop the error returned by the adapted.
	}

	vs := make([]interface{}, numOut)

	for i := 0; i < numOut; i++ {
		vs[i] = r[i].Interface()
	}

	p := a.findTypeProcessorOr(SuccessfullyInvokedType, defaultSuccessProcessor)
	_, _ = p(c, vs...)

	return nil
}

func (a *Adaptor) createArgs(c *gin.Context, fv reflect.Value, option *Option) ([]reflect.Value, error) {
	ft := fv.Type()
	numIn := ft.NumIn()

	ii := -1
	argVs := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		argType, argKind, isArgTypePtr := parseArgs(ft, i)
		switch argKind {
		case reflect.Struct:
			v, err := a.processStruct(c, argType)
			if err != nil {
				return nil, err
			}

			argVs[i] = convertPtr(isArgTypePtr, v)
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

func (a *Adaptor) processStruct(c *gin.Context, argType reflect.Type) (reflect.Value, error) {
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
