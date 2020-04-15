package giu

import (
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bingoohuang/gor"
	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
)

// TypeProcessor is the processor for a specified type.
type TypeProcessor func(*gin.Context, ...interface{}) (interface{}, error)

// Processor is the processor for a specified type.
type Processor func(*gin.Context, ...interface{})

// Adaptor is the adaptor structure for gin.HandlerFunc.
type Adaptor struct {
	typeProcessors    map[reflect.Type]TypeProcessor
	arounderFactories map[string]InvokeArounderFactory
	arounders         map[reflect.Type]InvokeArounder
	router            *httprouter.Router
}

// NewAdaptor makes a new Adaptor.
func NewAdaptor() *Adaptor {
	a := &Adaptor{
		typeProcessors:    make(map[reflect.Type]TypeProcessor),
		arounderFactories: make(map[string]InvokeArounderFactory),
		arounders:         make(map[reflect.Type]InvokeArounder),
		router:            httprouter.New(),
	}

	a.RegisterTypeProcessor(reflect.TypeOf((*DownloadFile)(nil)).Elem(), downloadFileProcessor)
	a.RegisterTypeProcessor(reflect.TypeOf((*DirectResponse)(nil)).Elem(), directResponseProcessor)

	return a
}

// RegisterInvokeArounder register arounder for the adaptor.
func (a *Adaptor) RegisterInvokeArounder(arounderName string, arounder InvokeArounderFactory) {
	a.arounderFactories[arounderName] = arounder
}

// RegisterErrProcessor typeProcessors a type processor for the error.
func (a *Adaptor) RegisterErrProcessor(p Processor) {
	a.typeProcessors[gor.ErrType] = func(c *gin.Context, args ...interface{}) (interface{}, error) {
		p(c, args...)
		return nil, nil
	}
}

// RegisterSuccProcessor typeProcessors a type processor for the successful deal.
func (a *Adaptor) RegisterSuccProcessor(p Processor) {
	a.typeProcessors[NonPtrTypeOf(SuccInvokedType)] =
		func(c *gin.Context, args ...interface{}) (interface{}, error) {
			p(c, args...)
			return nil, nil
		}
}

// RegisterTypeProcessor typeProcessors a type processor for the type.
func (a *Adaptor) RegisterTypeProcessor(t interface{}, p TypeProcessor) {
	a.typeProcessors[NonPtrTypeOf(t)] = p
}

func (a *Adaptor) findProcessor(v interface{}) TypeProcessor {
	src := reflect.TypeOf(v)

	for t, p := range a.typeProcessors {
		if gor.ImplType(src, t) {
			return p
		}
	}

	return nil
}

func (a *Adaptor) findTypeProcessor(t reflect.Type) TypeProcessor {
	for k, v := range a.typeProcessors {
		if gor.ImplType(t, k) {
			return v
		}
	}

	return nil
}

func (a *Adaptor) findTypeProcessorOr(t reflect.Type, processor TypeProcessor) TypeProcessor {
	if p := a.findTypeProcessor(t); p != nil {
		return p
	}

	return processor
}

// succInvoked represents the adaptor is invoked successfully with returned value.
type succInvoked interface{ succInvoked() }

// SuccInvokedType defines the successfully adaptor invoke's type.
// nolint gochecknoglobals
var SuccInvokedType = reflect.TypeOf((*succInvoked)(nil)).Elem()

// NonPtrTypeOf returns the non-ptr type of v.
func NonPtrTypeOf(v interface{}) reflect.Type {
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

// JSONValuer defines the interface of how to convert a JSON object.
type JSONValuer interface {
	// JSONValue converts the current object to an object.
	JSONValue() (interface{}, error)
}

// HTTPStatus defines the type of HTTP state.
type HTTPStatus int

func defaultSuccProcessor(g *gin.Context, vs ...interface{}) (interface{}, error) {
	if len(vs) > 0 {
		_ = defaultSuccProcessorInternal(&ginContext{g}, vs...)
	}

	return nil, nil
}

// responder responds the http request.
type responder interface {
	Status(code int) error
	JSON(code int, obj interface{}) error
	String(code int, format string, values ...interface{}) error
}

type ginContext struct {
	*gin.Context
}

func (g *ginContext) Status(code int) error                { g.Context.Status(code); return nil }
func (g *ginContext) JSON(code int, obj interface{}) error { g.Context.JSON(code, obj); return nil }
func (g *ginContext) String(code int, format string, values ...interface{}) error {
	g.Context.String(code, format, values...)
	return nil
}

func defaultSuccProcessorInternal(g responder, vs ...interface{}) error {
	code, vs := findStateCode(vs)

	if len(vs) == 0 {
		return g.Status(code)
	}

	if found, err := findJSONValuer(g, vs, code); found {
		return err
	}

	if len(vs) == 1 { // nolint gomnd
		return respondOut1(g, vs, code)
	}

	m := make(map[string]interface{})

	for _, v := range vs {
		m[reflect.TypeOf(v).String()] = v
	}

	return g.JSON(code, m)
}

func respondOut1(g responder, vs []interface{}, code int) error {
	switch v0 := vs[0]; reflect.Indirect(reflect.ValueOf(v0)).Kind() {
	case reflect.Struct, reflect.Map:
		return g.JSON(code, v0)
	default:
		return g.String(code, "%v", v0)
	}
}

func findJSONValuer(g responder, vs []interface{}, code int) (bool, error) {
	for _, v := range vs {
		if vj, ok := v.(JSONValuer); ok {
			jv, err := vj.JSONValue()
			if err != nil {
				return true, err
			}

			return true, g.JSON(code, jv)
		}
	}

	return false, nil
}

func findStateCode(vs []interface{}) (int, []interface{}) {
	code := http.StatusOK
	vvs := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		if vv, ok := v.(HTTPStatus); ok {
			code = int(vv)
		} else {
			vvs = append(vvs, v)
		}
	}

	return code, vvs
}

func defaultErrorProcessor(g *gin.Context, vs ...interface{}) (interface{}, error) {
	code := 500
	if sce, ok := vs[0].(StateCodeError); ok {
		code = sce.GetStateCode()
	}

	_ = g.AbortWithError(code, vs[0].(error))

	return nil, nil
}

// ExpandableParam defines the interface that the param can be expanded to multiples.
type ExpandableParam interface {
	// Expands expands one item to multiple items.
	Expands() []Param
}

// Param defines the interface to how to get param's value.
type Param interface {
	Get(g *gin.Context) string
}

type urlParams struct {
	keys []string
}

// Expands expands one item to multiple items.
func (u urlParams) Expands() []Param {
	params := make([]Param, len(u.keys))

	for i, key := range u.keys {
		params[i] = urlParam{key: key}
	}

	return params
}

type urlParam struct{ key string }

// URLParams defines the URL param in the URL PATH.
func URLParams(keys ...string) ExpandableParam { return urlParams{keys: keys} }

func (u urlParam) Get(g *gin.Context) string { return g.Param(u.key) }

var _ Param = (*urlParam)(nil)

type queryParams struct{ keys []string }

func (q queryParams) Expands() []Param {
	params := make([]Param, len(q.keys))

	for i, key := range q.keys {
		params[i] = queryParam{key: key}
	}

	return params
}

type queryParam struct {
	key          string
	defaultValue string
	required     bool
}

func (u queryParam) Expands() []Param {
	return []Param{u}
}

// QueryParams defines the query param.
func QueryParams(keys ...string) ExpandableParam {
	return queryParams{keys: keys}
}

// QueryParamOr defines the query param and the default value when empty.
func QueryParamOr(key, defaultValue string, required bool) ExpandableParam {
	return queryParam{key: key, defaultValue: defaultValue, required: required}
}

func (u queryParam) Get(g *gin.Context) string {
	return g.DefaultQuery(u.key, u.defaultValue)
}

var _ Param = (*queryParam)(nil)

// Option defines the adaptor's option.
type Option struct {
	Params     []Param
	MiddleWare bool
}

// OptionFn is the function prototype to apply option.
type OptionFn func(*Option)

// Params defines the params for the adaptor.
func Params(ps ...ExpandableParam) OptionFn {
	params := make([]Param, 0)

	for _, p := range ps {
		params = append(params, p.Expands()...)
	}

	return func(option *Option) { option.Params = params }
}

// MiddleWare defines the middleWare flag for the adaptor.
func MiddleWare(m bool) OptionFn { return func(option *Option) { option.MiddleWare = m } }

// Adapt adapts convenient function to gi.HandleFunc.
func (a *Adaptor) Adapt(fn HandlerFunc, optionFns ...OptionFn) gin.HandlerFunc {
	fv := reflect.ValueOf(fn)
	option := a.makeOption(optionFns)
	errTp := a.findTypeProcessorOr(gor.ErrType, defaultErrorProcessor)

	return func(c *gin.Context) {
		if err := a.internalAdapter(c, fv, option); err != nil {
			_, _ = errTp(c, err)
		}
	}
}

func (a *Adaptor) internalAdapter(c *gin.Context, fv reflect.Value, option *Option) error {
	argVs, err := a.createArgs(c, fv, option)
	if err != nil {
		return err
	}

	fa := a.arounders[fv.Type()]

	if err := around(fa, argVs, true); err != nil {
		return err
	}

	r := fv.Call(argVs)

	_ = around(fa, r, false)

	return a.processOut(c, fv, r, option)
}

func (a *Adaptor) makeOption(optionFns []OptionFn) *Option {
	option := &Option{}

	for _, f := range optionFns {
		f(option)
	}

	return option
}

func around(fa InvokeArounder, v []reflect.Value, before bool) error {
	if fa == nil {
		return nil
	}

	args := make([]interface{}, len(v))
	for i, a := range v {
		args[i] = a.Interface()
	}

	if before {
		return fa.Before(args)
	}

	fa.After(args)

	return nil
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

	a.registerInjects(c, option, numOut, r)

	if option.MiddleWare {
		return nil
	}

	a.succProcess(c, numOut, r)

	return nil
}

func (a *Adaptor) succProcess(c *gin.Context, numOut int, r []reflect.Value) {
	vs := make([]interface{}, numOut)

	for i := 0; i < numOut; i++ {
		vs[i] = r[i].Interface()
	}

	if numOut > 0 {
		if tp := a.findProcessor(vs[0]); tp != nil {
			_, _ = tp(c, vs...)
			return
		}
	}

	p := a.findTypeProcessorOr(SuccInvokedType, defaultSuccProcessor)
	_, _ = p(c, vs...)
}

func (a *Adaptor) registerInjects(c *gin.Context, option *Option, numOut int, r []reflect.Value) {
	if !option.MiddleWare {
		return
	}

	for i := 0; i < numOut; i++ {
		nonPtrRi := convertPtr(false, r[i])
		if nonPtrRi.Kind() == reflect.Struct {
			c.Set(a.injectKey(nonPtrRi.Type()), convertPtr(true, r[i]))
		}
	}
}

type argIn struct {
	Index          int
	Type           reflect.Type
	Kind           reflect.Kind
	Ptr            bool
	PrimitiveIndex int
}

func (a *Adaptor) createArgs(c *gin.Context, fv reflect.Value, option *Option) (v []reflect.Value, err error) {
	ft := fv.Type()
	argIns := parseArgIns(ft)
	argTags := collectTags(argIns)
	argValuesByTag := argTags.createArgValues(c)
	pArg := singlePrimitiveValue(c, argTags.countPrimitiveArgs(argIns))

	v = make([]reflect.Value, ft.NumIn())

	for i, arg := range argIns {
		if v[i], err = a.createArgValue(c, argValuesByTag, argTags, arg, pArg, option); err != nil {
			return nil, err
		}
	}

	return v, err
}

func singlePrimitiveValue(c *gin.Context, primitiveArgsNum int) string {
	if primitiveArgsNum != 1 { // nolint gomnd
		return ""
	}

	if len(c.Params) == 1 { // nolint gomnd
		return c.Params[0].Value
	}

	q := c.Request.URL.Query()
	if len(q) == 1 { // nolint gomnd
		for _, v := range q {
			return v[0]
		}
	}

	return ""
}

func (t ArgsTags) countPrimitiveArgs(argIns []argIn) int {
	primitiveArgsNum := 0

	for i, arg := range argIns {
		if _, ok := t[i]; ok || arg.Kind == reflect.Struct {
			continue
		}

		argIns[i].PrimitiveIndex = primitiveArgsNum
		primitiveArgsNum++
	}

	return primitiveArgsNum
}

func (a *Adaptor) createArgValue(c *gin.Context, argValuesByTag map[int]string,
	argAsTags ArgsTags, arg argIn, singleArgValue string, option *Option) (reflect.Value, error) {
	if _, ok := argAsTags[arg.Index]; ok {
		return convertPtr(arg.Ptr, reflect.New(arg.Type)), nil
	}

	if arg.Kind == reflect.Struct {
		v, err := a.processStruct(c, arg)
		if err != nil {
			return reflect.Value{}, err
		}

		return convertPtr(arg.Ptr, v), nil
	}

	if arg.PrimitiveIndex < 0 {
		return reflect.Value{}, fmt.Errorf("unable to parse arg%d for %s", arg.Index, arg.Type)
	}

	if v, ok := argValuesByTag[arg.PrimitiveIndex]; ok {
		return convertValue(v, arg)
	}

	if arg.PrimitiveIndex < len(option.Params) {
		v := option.Params[arg.PrimitiveIndex].Get(c)
		return convertValue(v, arg)
	}

	if singleArgValue != "" {
		return convertValue(singleArgValue, arg)
	}

	return reflect.Zero(arg.Type), nil
}

func convertValue(singleArgValue string, arg argIn) (reflect.Value, error) {
	v, err := gor.CastAny(singleArgValue, arg.Type)
	if err != nil {
		return reflect.Value{}, err
	}

	return convertPtr(arg.Ptr, v), nil
}

func parseArgIns(ft reflect.Type) []argIn {
	numIn := ft.NumIn()
	argIns := make([]argIn, numIn)

	for i := 0; i < numIn; i++ {
		argIns[i] = parseArgs(ft, i)
	}

	return argIns
}

func (t ArgsTags) createArgValues(c *gin.Context) map[int]string {
	args := map[int]string{}

	t.collectTagValues("arg", func(v string) bool {
		for _, argItem := range strings.Split(v, "/") {
			parseTags(c, argItem, args)
		}

		return false
	})

	return args
}

func (t ArgsTags) getFirstTagValues(tagName string) (v string, ok bool) {
	t.collectTagValues(tagName, func(tag string) bool {
		v = tag
		ok = true

		return true
	})

	return
}

func (t ArgsTags) collectTagValues(tagName string, fn func(string) bool) {
	for _, tags := range t {
		for _, tag := range tags {
			if v, ok := tag.Lookup(tagName); ok && fn(v) {
				return
			}
		}
	}
}

func parseTags(c *gin.Context, arg string, args map[int]string) {
	parts := strings.Split(arg, ",")
	namesStr, mode := parts[0], parts[1]

	for _, name := range strings.Fields(namesStr) {
		switch mode {
		case "url":
			args[len(args)] = c.Param(name)
		case "query":
			args[len(args)] = c.Query(name)
		case "form":
			args[len(args)], _ = c.GetPostForm(name)
		case "context":
			args[len(args)] = c.GetString(name)
		}
	}
}

type ArgsTags map[int][]reflect.StructTag

func collectTags(args []argIn) ArgsTags {
	argTags := make(ArgsTags)

	for i, arg := range args {
		if arg.Kind != reflect.Struct {
			continue
		}

		if tags := findTags(arg.Type, TType); len(tags) > 0 {
			argTags[i] = tags
		}
	}

	return argTags
}

func findTags(t reflect.Type, target reflect.Type) []reflect.StructTag {
	tags := make([]reflect.StructTag, 0)

	for i := 0; i < t.NumField(); i++ {
		if tf := t.Field(i); tf.Type == target {
			tags = append(tags, tf.Tag)
		}
	}

	return tags
}

func parseArgs(ft reflect.Type, argIndex int) argIn {
	argType := ft.In(argIndex)
	ptr := argType.Kind() == reflect.Ptr

	if ptr {
		argType = argType.Elem()
	}

	return argIn{Index: argIndex, Type: argType, Kind: argType.Kind(), Ptr: ptr, PrimitiveIndex: -1}
}

func (a *Adaptor) processStruct(c *gin.Context, arg argIn) (reflect.Value, error) {
	if arg.Ptr && arg.Type == NonPtrTypeOf(c) { // 直接注入gin.Context
		return reflect.ValueOf(c), nil
	}

	if v, exists := c.Get(a.injectKey(arg.Type)); exists {
		return convertPtr(arg.Ptr, v.(reflect.Value)), nil
	}

	if tp := a.findTypeProcessor(arg.Type); tp != nil {
		v, err := tp(c, arg.Type)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(v), nil
	}

	argValue := reflect.New(arg.Type)
	if err := c.ShouldBind(argValue.Interface()); err != nil {
		return reflect.Value{}, err
	}

	return argValue, nil
}

func (a *Adaptor) injectKey(t reflect.Type) string { return "_inject_" + t.String() }

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
	a.GinRouter.Handle(httpMethod, relativePath, a.Adaptor.Adapt(h, fns...))
	return a
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (a *Routes) POST(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("POST", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (a *Routes) GET(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("GET", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (a *Routes) DELETE(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("DELETE", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (a *Routes) PATCH(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("PATCH", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (a *Routes) PUT(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("PUT", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (a *Routes) OPTIONS(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("OPTIONS", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (a *Routes) HEAD(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Handle("HEAD", relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (a *Routes) Any(relativePath string, h HandlerFunc, optionFns ...OptionFn) IRoutes {
	a.GinRouter.Any(relativePath, a.Adaptor.Adapt(h, optionFns...))
	return a
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc interface{}

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

	HandleFn(...HandlerFunc) IRoutes
}

// Route makes a route for Adaptor.
func (a *Adaptor) Route(r gin.IRouter) IRoutes { return &Routes{GinRouter: r, Adaptor: a} }

// Keep keeps the URL related data.
type Keep struct {
	Path    string
	Keep    string
	Methods []string
}

type keepResponseWriter struct {
	http.ResponseWriter
	keep *Keep
}

// FindKeep finds the keep data for the current request.
func (a *Adaptor) FindKeep(c *gin.Context) *Keep {
	kw := &keepResponseWriter{ResponseWriter: httptest.NewRecorder()}
	a.router.ServeHTTP(kw, c.Request)

	return kw.keep
}

func (a *Adaptor) keep(methods []string, hasAny bool, absolutePath, keep string) {
	f := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if kw, ok := w.(*keepResponseWriter); ok {
			kw.keep = &Keep{Path: absolutePath, Keep: keep, Methods: methods}
		}
	}

	if hasAny {
		a.any(absolutePath, f)
	} else {
		for _, m := range methods {
			a.router.Handle(m, absolutePath, f)
		}
	}
}

// any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (a *Adaptor) any(relativePath string, handlers httprouter.Handle) {
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS", "DELETE", "CONNECT", "TRACE"} {
		a.router.Handle(m, relativePath, handlers)
	}
}

// Routes defines adaptor routes implemetation for IRoutes.
type Routes struct {
	GinRouter gin.IRouter
	Adaptor   *Adaptor
}

// HandleFn will typeProcessors h by its declaration.
func (a *Routes) HandleFn(hs ...HandlerFunc) IRoutes {
	for _, h := range hs {
		ht := reflect.TypeOf(h)
		if ht.Kind() == reflect.Func {
			a.handleFn("", h, false)
			continue
		}

		if ht.Kind() == reflect.Ptr {
			ht = ht.Elem()
		}

		if ht.Kind() == reflect.Interface {
			panic(fmt.Errorf("invalid handler type %v for %v", ht, h))
		}

		v := reflect.ValueOf(h)

		for i := 0; i < ht.NumMethod(); i++ {
			a.handleFn(ht.Method(i).Name, v.Method(i).Interface(), true)
		}
	}

	return a
}

func (a *Routes) handleFn(handlerName string, h HandlerFunc, ignoreIllegal bool) {
	ht := reflect.TypeOf(h)
	tags := collectTags(parseArgIns(ht))

	for rounderName, factory := range a.Adaptor.arounderFactories {
		if v, ok := tags.getFirstTagValues(rounderName); ok {
			a.Adaptor.arounders[ht] = factory.Create(handlerName, v, h)
		}
	}

	url, _ := tags.getFirstTagValues("url")

	if url = strings.TrimSpace(url); url == "" {
		if ignoreIllegal {
			return
		}

		panic("unable to find url")
	}

	methods, hasAny, relativePath := a.parseMethodRelativePath(url)

	a.registerKeep(tags, relativePath, methods, hasAny)

	if hasAny {
		a.GinRouter.Any(relativePath, a.Adaptor.Adapt(h))
	} else {
		for _, method := range methods {
			a.GinRouter.Handle(method, relativePath, a.Adaptor.Adapt(h))
		}
	}
}

func (a *Routes) registerKeep(tags ArgsTags, relativePath string, methods []string, hasAny bool) {
	if keep, _ := tags.getFirstTagValues("keep"); keep != "" {
		type basePather interface{ BasePath() string }

		if bp, ok := a.GinRouter.(basePather); ok {
			absolutePath := joinPaths(bp.BasePath(), relativePath)
			a.Adaptor.keep(methods, hasAny, absolutePath, keep)
		}
	}
}

// copied from github.com/gin-gonic/gin@v1.5.0/utils.go
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'

	if appendSlash {
		return finalPath + "/"
	}

	return finalPath
}

func lastChar(str string) uint8 {
	if str != "" {
		return str[len(str)-1]
	}

	panic("The length of the string can't be 0")
}

const anyMethod = "ANY"

func (a *Routes) parseMethodRelativePath(url string) ([]string, bool, string) {
	urlFields := strings.Fields(url)

	// 没有定义HTTP METHOD，当做ANY
	if len(urlFields) == 1 { // nolint gomnd
		return []string{anyMethod}, true, urlFields[0]
	}

	allMethodsStr := strings.ToUpper(urlFields[0])
	allMethods := strings.Split(allMethodsStr, "/")
	methodMap := make(map[string]bool)
	hasAny := false

	for _, m := range allMethods {
		if m == anyMethod {
			hasAny = true
			break
		}

		methodMap[m] = true
	}

	if hasAny {
		return []string{anyMethod}, true, urlFields[1]
	}

	methods := make([]string, 0, len(methodMap))
	for k := range methodMap {
		methods = append(methods, k)
	}

	return methods, false, urlFields[1]
}

// Use adds middleware, see example code.
func (a *Routes) Use(h HandlerFunc, optionFns ...OptionFn) IRoutes {
	fns := make([]OptionFn, len(optionFns)+1) // nolint gomnd
	copy(fns, optionFns)
	fns[len(optionFns)] = MiddleWare(true)
	a.GinRouter.Use(a.Adaptor.Adapt(h, fns...))

	return a
}

var _ IRoutes = (*Routes)(nil)

// T defines the tag for handler functions.
type T interface{ t() }

var (
	// TType defines the type of T.
	TType = reflect.TypeOf((*T)(nil)).Elem() // nolint gochecknoglobals
)

// InvokeArounderFactory defines the factory to create InvokeArounder
type InvokeArounderFactory interface {
	// Create creates the InvokeArounder with the tag value and the handler type.
	Create(handlerName, tag string, handler HandlerFunc) InvokeArounder
}

// InvokeArounder defines the adaptee invoking before and after intercepting points for the user.
type InvokeArounder interface {
	// Before will be called before the adaptee invoking.
	Before(args []interface{}) error

	// After will be called after the adaptee invoking.
	After(outs []interface{})
}

// DirectResponse represents the direct response.
type DirectResponse struct {
	Code  int
	Error error
}

// directResponseProcessor is the processor for DirectResponse.
func directResponseProcessor(c *gin.Context, args ...interface{}) (interface{}, error) {
	dr, ok := args[0].(*DirectResponse)
	if !ok {
		arg0 := args[0].(DirectResponse)
		dr = &arg0
	}

	if dr.Code == 0 {
		dr.Code = http.StatusOK
	}

	c.AbortWithStatus(dr.Code)

	if dr.Error != nil {
		_ = c.Error(dr.Error)
	}

	return nil, nil
}

// DownloadFile represents the file to be downloaded.
type DownloadFile struct {
	DiskFile string
	Filename string
	Content  []byte
}

// downloadFileProcessor is the processor for a specified type.
func downloadFileProcessor(c *gin.Context, args ...interface{}) (interface{}, error) {
	df, ok := args[0].(*DownloadFile)
	if !ok {
		arg0 := args[0].(DownloadFile)
		df = &arg0
	}

	cd := createContentDisposition(df)

	c.Header("Content-Disposition", cd)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	if df.DiskFile != "" {
		c.File(df.DiskFile)
		return nil, nil
	}

	_, _ = c.Writer.Write(df.Content)

	return nil, nil
}

func createContentDisposition(downloadFile *DownloadFile) string {
	m := map[string]string{"filename": getDownloadFilename(downloadFile)}
	return mime.FormatMediaType("attachment", m)
}

func getDownloadFilename(downloadFile *DownloadFile) string {
	filename := downloadFile.Filename

	if filename == "" {
		filename = filepath.Base(downloadFile.DiskFile)
	}

	if filename == "" {
		return "dl"
	}

	return filename
}
