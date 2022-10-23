package app

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type SelfCfg struct {
	Scheme string `yaml:"scheme"`
	Port   int    `yaml:"port"`
}

type consulCfg struct {
	Scheme  string `yaml:"scheme"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type config struct {
	Directories []string  `yaml:"directories"`
	FileTypes   []string  `yaml:"fileTypes"`
	Consul      consulCfg `yaml:"consul"`
	SelfCfg     `yaml:"mediaHost"`
}

func readConfig() (s config, err error) {
	var configPath string
	flag.StringVar(&configPath, "config", "", "optional path to config file")

	flag.Parse()

	if configPath == "" {
		cwd, err := os.Getwd()

		if err != nil {
			return s, err
		}

		configPath = cwd + "/config.yaml"
	}

	f, err := os.ReadFile(configPath)

	if err != nil {
		return
	}

	err = yaml.Unmarshal(f, &s)

	return
}
