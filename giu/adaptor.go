package giu

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"strconv"
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
	return &Adaptor{
		typeProcessors:    make(map[reflect.Type]TypeProcessor),
		arounderFactories: make(map[string]InvokeArounderFactory),
		arounders:         make(map[reflect.Type]InvokeArounder),
		router:            httprouter.New(),
	}
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
	a.typeProcessors[GetNonPtrType(SuccInvokedType)] = func(c *gin.Context, args ...interface{}) (interface{}, error) {
		p(c, args...)
		return nil, nil
	}
}

// RegisterTypeProcessor typeProcessors a type processor for the type.
func (a *Adaptor) RegisterTypeProcessor(t interface{}, p TypeProcessor) {
	a.typeProcessors[GetNonPtrType(t)] = p
}

func (a *Adaptor) findTypeProcessor(t reflect.Type) TypeProcessor {
	for k, v := range a.typeProcessors {
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

// JSONValuer defines the interface of how to convert a JSON object.
type JSONValuer interface {
	// JSONValue converts the current object to an object.
	JSONValue() (interface{}, error)
}

func defaultSuccessProcessor(g *gin.Context, vs ...interface{}) (interface{}, error) {
	if len(vs) == 0 {
		return nil, nil
	}

	v0 := vs[0]

	if vj, ok := v0.(JSONValuer); ok {
		jv, err := vj.JSONValue()
		if err != nil {
			return nil, err
		}

		g.JSON(http.StatusOK, jv)

		return nil, nil
	}

	if reflect.Indirect(reflect.ValueOf(v0)).Kind() == reflect.Struct {
		g.JSON(http.StatusOK, v0)
	} else {
		g.String(http.StatusOK, fmt.Sprintf("%v", v0))
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

type urlParam struct {
	key string
}

// URLParams defines the URL param in the URL PATH.
func URLParams(keys ...string) ExpandableParam {
	return urlParams{keys: keys}
}

func (u urlParam) Get(g *gin.Context) string {
	return g.Param(u.key)
}

var _ Param = (*urlParam)(nil)

type queryParams struct {
	keys []string
}

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

// Option defines the adatpor's option.
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

	return func(option *Option) {
		option.Params = params
	}
}

// MiddleWare defines the middleWare flag for the adaptor.
func MiddleWare(m bool) OptionFn { return func(option *Option) { option.MiddleWare = m } }

// Adapt adapts convenient function to gi.HandleFunc.
func (a *Adaptor) Adapt(fn HandlerFunc, optionFns ...OptionFn) gin.HandlerFunc {
	option := a.makeOption(optionFns)
	fv := reflect.ValueOf(fn)
	errTp := a.findTypeProcessorOr(gor.ErrType, defaultErrorProcessor)

	return func(c *gin.Context) {
		if err := a.internalAdatpr(c, fv, option); err != nil {
			_, _ = errTp(c, err)
		}
	}
}

func (a *Adaptor) internalAdatpr(c *gin.Context, fv reflect.Value, option *Option) error {
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
	numIn := ft.NumIn()
	argIns := parseArgIns(ft)
	argAsTags := collectTags(argIns)
	argValuesByTag := createArgValues(c, argAsTags)
	primitiveArgsNum := countPrimitiveArgs(argIns, argAsTags)
	pArg := singlePrimitiveValue(c, primitiveArgsNum)

	v = make([]reflect.Value, numIn)

	for i, arg := range argIns {
		if v[i], err = a.createArgValue(c, argValuesByTag, argAsTags, arg, pArg, option); err != nil {
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

func countPrimitiveArgs(argIns []argIn, argAsTags map[int][]reflect.StructTag) int {
	primitiveArgsNum := 0

	for i, arg := range argIns {
		if _, ok := argAsTags[i]; ok || arg.Kind == reflect.Struct {
			continue
		}

		argIns[i].PrimitiveIndex = primitiveArgsNum
		primitiveArgsNum++
	}

	return primitiveArgsNum
}

func (a *Adaptor) createArgValue(c *gin.Context, argValuesByTag map[int]string,
	argAsTags map[int][]reflect.StructTag, arg argIn, singleArgValue string, option *Option) (reflect.Value, error) {
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

	return reflect.Value{}, fmt.Errorf("unable to parse arg%d for %s", arg.Index, arg.Type)
}

func convertValue(singleArgValue string, arg argIn) (reflect.Value, error) {
	v, err := dealDirectParamArg(singleArgValue, arg.Kind)
	if err != nil {
		return reflect.Value{}, err
	}

	return convertPtr(arg.Ptr, reflect.ValueOf(v)), nil
}

func parseArgIns(ft reflect.Type) []argIn {
	numIn := ft.NumIn()
	argIns := make([]argIn, numIn)

	for i := 0; i < numIn; i++ {
		argIns[i] = parseArgs(ft, i)
	}

	return argIns
}

func createArgValues(c *gin.Context, argTags map[int][]reflect.StructTag) map[int]string {
	args := map[int]string{}

	collectTagValues(argTags, "arg", func(v string) bool {
		for _, argItem := range strings.Split(v, "/") {
			parseTags(c, argItem, args)
		}

		return false
	})

	return args
}

func getFirstTagValues(argTags map[int][]reflect.StructTag, tagName string) (v string, ok bool) {
	collectTagValues(argTags, tagName, func(tag string) bool {
		v = tag
		ok = true

		return true
	})

	return
}
func collectTagValues(argTags map[int][]reflect.StructTag, tagName string, fn func(string) bool) {
	for _, tags := range argTags {
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

func collectTags(args []argIn) map[int][]reflect.StructTag {
	argTags := make(map[int][]reflect.StructTag)

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
		tf := t.Field(i)
		if tf.Type == target {
			tags = append(tags, tf.Tag)
		}
	}

	return tags
}

func dealDirectParamArg(argValue string, argKind reflect.Kind) (interface{}, error) {
	switch argKind {
	case reflect.String:
		return argValue, nil
	case reflect.Int:
		return strconv.Atoi(argValue)
	}

	return nil, fmt.Errorf("unsupported type %v", argKind)
}

func parseArgs(ft reflect.Type, argIndex int) argIn {
	argType := ft.In(argIndex)
	ptr := argType.Kind() == reflect.Ptr

	if ptr {
		argType = argType.Elem()
	}

	return argIn{
		Index:          argIndex,
		Type:           argType,
		Kind:           argType.Kind(),
		Ptr:            ptr,
		PrimitiveIndex: -1,
	}
}

func (a *Adaptor) processStruct(c *gin.Context, arg argIn) (reflect.Value, error) {
	if arg.Ptr && arg.Type == GetNonPtrType(c) { // 直接注入gin.Context
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

	HandleFn(...HandlerFunc) IRoutes
}

// Route makes a route for Adaptor.
func (a *Adaptor) Route(r gin.IRouter) IRoutes { return &Routes{GinRouter: r, Adaptor: a} }

type Keep struct {
	Path    string
	Keep    string
	Methods []string
}

type KeepResponseWriter struct {
	http.ResponseWriter

	keep *Keep
}

func (a *Adaptor) FindKeep(c *gin.Context) *Keep {
	kw := &KeepResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
	}

	a.router.ServeHTTP(kw, c.Request)

	return kw.keep
}

func (a *Adaptor) keep(methods []string, hasAny bool, absolutePath, keep string) {
	f := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if kw, ok := w.(*KeepResponseWriter); ok {
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
	a.router.Handle("GET", relativePath, handlers)
	a.router.Handle("POST", relativePath, handlers)
	a.router.Handle("PUT", relativePath, handlers)
	a.router.Handle("PATCH", relativePath, handlers)
	a.router.Handle("HEAD", relativePath, handlers)
	a.router.Handle("OPTIONS", relativePath, handlers)
	a.router.Handle("DELETE", relativePath, handlers)
	a.router.Handle("CONNECT", relativePath, handlers)
	a.router.Handle("TRACE", relativePath, handlers)
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
		f := factory

		if v, ok := getFirstTagValues(tags, rounderName); ok {
			a.Adaptor.arounders[ht] = f.Create(handlerName, v, h)
		}
	}

	url, _ := getFirstTagValues(tags, "url")

	if url = strings.TrimSpace(url); url == "" {
		if ignoreIllegal {
			return
		}

		panic("unable to find url")
	}

	methods, hasAny, relativePath := a.parseMethodRelativePath(url)

	if keep, _ := getFirstTagValues(tags, "keep"); keep != "" {
		type basePather interface{ BasePath() string }

		if bp, ok := a.GinRouter.(basePather); ok {
			absolutePath := joinPaths(bp.BasePath(), relativePath)
			a.Adaptor.keep(methods, hasAny, absolutePath, keep)
		}
	}

	if hasAny {
		a.GinRouter.Any(relativePath, a.Adaptor.Adapt(h))
	} else {
		for _, method := range methods {
			a.GinRouter.Handle(method, relativePath, a.Adaptor.Adapt(h))
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
	if str == "" {
		panic("The length of the string can't be 0")
	}

	return str[len(str)-1]
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
