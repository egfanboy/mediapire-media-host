package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Directories []string `yaml:"directories"`
	FileTypes   []string `yaml:"fileTypes"`
}

func (c Config) IsMediaSupported(extension string) bool {

	for _, t := range c.FileTypes {
		if extension == t {
			return true
		}
	}

	return false
}

func readConfig() (s Config, err error) {
	path, err := os.Getwd()

	if err != nil {
		return
	}
	f, err := os.ReadFile(path + "/../config.yaml")

	if err != nil {
		return
	}

	err = yaml.Unmarshal(f, &s)

	return
}
