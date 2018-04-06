package paths

import (
	"os"
	"os/user"
	"strings"
)

// GetHome ...
func GetHome() (string, error) {
	var home string
	home = os.Getenv("HOME")
	if home == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		home = u.HomeDir
	}
	return home, nil
}

// ExpandUser ...
func ExpandUser(path string) (string, error) {
	if len(path) > 0 && strings.HasPrefix(path, "~") {
		home, err := GetHome()
		if err != nil {
			return "", err
		}
		return strings.Replace(path, "~", home, 1), nil
	}
	return path, nil
}
