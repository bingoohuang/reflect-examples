package giu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type myResponder struct {
	code   int
	obj    interface{}
	format string
	values []interface{}
}

func (g *myResponder) Status(code int) error {
	g.code = code

	return nil
}

func (g *myResponder) JSON(code int, obj interface{}) error {
	g.code = code
	g.obj = obj

	return nil
}

func (g *myResponder) String(code int, format string, values ...interface{}) error {
	g.code = code
	g.format = format
	g.values = values

	return nil
}

func TestRespondInternal(t *testing.T) {
	r := &myResponder{}

	assert.Nil(t, defaultSuccProcessorInternal(r))
	assert.Equal(t, 200, r.code)

	assert.Nil(t, defaultSuccProcessorInternal(r, HTTPStatus(123)))
	assert.Equal(t, 123, r.code)

	assert.Nil(t, defaultSuccProcessorInternal(r, HTTPStatus(123), "hello"))
	assert.Equal(t, "hello", r.values[0])
}
