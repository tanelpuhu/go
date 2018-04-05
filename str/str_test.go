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

func TestIsSeven(t *testing.T) {
	if !IsSeven("7") {
		t.Errorf("7 is Seven!")
	}
	if IsSeven("77") {
		t.Errorf("77 is not Seven!")
	}
}
