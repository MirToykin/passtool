package config

import "path/filepath"

func ensureTrailingSlash(path string) string {
	cleanedPath := filepath.Clean(path)
	if cleanedPath[len(cleanedPath)-1] != filepath.Separator {
		return cleanedPath + string(filepath.Separator)
	}
	return cleanedPath
}
