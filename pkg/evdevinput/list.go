package evdevinput

import (
	"io/ioutil"
	"path"

	evdev "github.com/holoplot/go-evdev"
)

func ListDevices() ([]string, error) {
	basePath := "/dev/input"

	result := make([]string, 0)

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return []string{}, err
	}

	for _, fileName := range files {
		if fileName.IsDir() {
			continue
		}

		fullPath := path.Join(basePath, fileName.Name())

		d, err := evdev.Open(fullPath)
		if err == nil {
			supportsKeys := false
			for _, evType := range d.CapableTypes() {
				if evType == evdev.EV_KEY {
					supportsKeys = true
					break
				}
			}

			if supportsKeys {
				result = append(result, fullPath)
			}
		}
	}
	return result, nil
}
