package examples

import (
	"fmt"
	"testing"

	"github.com/alangpierce/go-forceexport"
	"github.com/stretchr/testify/assert"
)

func TestForExport(t *testing.T) {
	var timeNow func() (int64, int32)
	err := forceexport.GetFunc(&timeNow, "time.now")
	assert.Nil(t, err)

	// Calls the actual time.now function.
	sec, nsec := timeNow()

	fmt.Println(sec, nsec)
}
