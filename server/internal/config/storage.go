package config

import "os"

const defaultDataDir = "data"

func DataDir() string {
	if dir := os.Getenv("APP_DATA_DIR"); dir != "" {
		return dir
	}

	return defaultDataDir
}
