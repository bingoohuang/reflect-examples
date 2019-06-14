package reflect_examples

import (
	"fmt"
	"testing"

	"github.com/alangpierce/go-forceexport"
)

func TestForExport(t *testing.T) {
	var timeNow func() (int64, int32)
	err := forceexport.GetFunc(&timeNow, "time.now")
	if err != nil {
		// Handle errors if you care about name possibly being invalid.
	}
	// Calls the actual time.now function.
	sec, nsec := timeNow()

	fmt.Println(sec, nsec)
}
