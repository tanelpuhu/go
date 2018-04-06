package paths

import (
	"strings"
	"testing"
)

func assertExpandUser(t *testing.T, val, exp string) {
	path, err := ExpandUser(val)
	if err != nil {
	}
	if path != exp {
		t.Errorf("%s -> %s != %s", val, path, exp)
	}
}

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

func TestExpandUser(t *testing.T) {
	home, _ := GetHome()
	assertExpandUser(t, "~", home)
	assertExpandUser(t, "~/", home+"/")
	assertExpandUser(t, "/tmp/", "/tmp/")
	assertExpandUser(t, "/var/www/~user", "/var/www/~user")
	assertExpandUser(t, "/var/www/~user/", "/var/www/~user/")
}
