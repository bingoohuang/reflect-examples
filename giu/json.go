package giu

import (
	"net/http"

	"github.com/bingoohuang/strcase"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// Jsonify renders JSON of obj to the response by the gin context.
func Jsonify(c *gin.Context, code int, obj interface{}) {
	c.Render(code, JSON{Data: obj})
}

// JSON contains the given interface object.
type JSON struct {
	Data interface{}
}

// Render (JSON) writes data with custom ContentType.
func (r JSON) Render(w http.ResponseWriter) error {
	return WriteJSON(w, r.Data)
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// nolint:gochecknoinits
func init() {
	extra.SetNamingStrategy(strcase.ToCamelLower)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	return jsoniter.NewEncoder(w).Encode(&obj)
}

// nolint
var (
	jsonContentType = []string{"application/json; charset=utf-8"}

	_ render.Render = JSON{}

	JSONUnmarshal     = jsoniter.Unmarshal
	JSONMarshal       = jsoniter.Marshal
	JSONMarshalIndent = jsoniter.MarshalIndent
)

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()

	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
