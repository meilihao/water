package water

import (
	"fmt"
	"testing"
)

func TestDefaultSerialId(t *testing.T) {
	var sa SerialAdapter = DefaultSerial{}
	id := sa.Id()

	if id == "" {
		t.Error("No Default Serial Id")
	} else {
		t.Log(fmt.Sprintf("Get Default Serial Id(%s)", id))
	}
}
