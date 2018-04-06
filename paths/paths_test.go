package paths

import (
	"strings"
	"testing"
)

func TestGetHome(t *testing.T) {
	home, err := GetHome()
	if err != nil {
		t.Errorf("Error was returned: %v", err)
	}

	if !strings.HasPrefix(home, "/") {
		t.Errorf("Home does not start with slash: %v", home)
	}

	if !strings.HasPrefix(home, "/home/") && !strings.HasPrefix(home, "/Users/") {
		// No Windows...
		t.Errorf("Weird home location: %v", home)
	}

}
