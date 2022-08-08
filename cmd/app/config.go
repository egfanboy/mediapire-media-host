package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SelfCfg struct {
	Scheme string `yaml:"scheme"`
	Port   int    `yaml:"port"`
}

type masterCfg struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

type config struct {
	Directories []string  `yaml:"directories"`
	FileTypes   []string  `yaml:"fileTypes"`
	Master      masterCfg `yaml:"master"`
	SelfCfg     `yaml:"mediaHost"`
}

func readConfig() (s config, err error) {
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
