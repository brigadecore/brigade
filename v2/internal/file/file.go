package file

import (
	"os"
)

// Exists returns a bool indicating if the specified file exists or not. It
// returns any errors that are encountered that are NOT an os.ErrNotExist error.
func Exists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// EnsureDirectory checks if the dirPath directory exists. If not, it
// creates it.
func EnsureDirectory(dirPath string) (bool, error) {
	if dirExists, err := Exists(dirPath); err != nil {
		return false, err
	} else if !dirExists {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
