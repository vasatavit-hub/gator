package config

import (
	"os"
)

func configPath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path + config_path, nil
}
