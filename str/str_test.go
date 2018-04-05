package str

import (
	"testing"
)

func TestIsStringIn(t *testing.T) {
	hay := []string{"1", "2", "3"}
	if !InSlice(hay, "1") {
		t.Errorf("Value 1 found...")
	}
	if InSlice(hay, "4") {
		t.Errorf("Value 4 found?!?!?!")
	}
}
