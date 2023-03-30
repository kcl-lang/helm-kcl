package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadfile(filepath string, url string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("download %s error, code: %d", url, resp.StatusCode)
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
