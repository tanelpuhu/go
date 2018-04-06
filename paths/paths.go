package paths

import (
	"os"
	"os/user"
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
