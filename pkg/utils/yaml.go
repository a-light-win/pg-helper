package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYaml(path string, v interface{}) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(bytes, v); err != nil {
		return err
	}

	return nil
}
