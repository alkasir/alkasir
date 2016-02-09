package clientconfig

import (
	"errors"
	"os"
	"path/filepath"
)

func mkConfigDir(subpaths ...string) (string, error) {
	dir := ConfigPath()
	if dir == "" {
		return "", os.ErrNotExist
	}
	args := make([]string, 0)
	args = append(args, dir)
	args = append(args, subpaths...)
	dir = filepath.Join(args...)

	exists, err := isDirExists(dir)
	if err != nil {
		return dir, err
	}
	if exists {
		return dir, nil
	}
	if err = os.Mkdir(dir, 0755); err != nil {
		return dir, errors.New("Error create config directory")
	}
	return dir, err
}

func isDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, errors.New(path + " exists but is not directory")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
