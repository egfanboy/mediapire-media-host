package app

import (
	"flag"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func getDownloadPath() (string, error) {
	basePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(basePath, ".mediapire", "mediahost", "downloads"), nil
}

type SelfCfg struct {
	Scheme  string  `yaml:"scheme"`
	Port    int     `yaml:"port"`
	Address *string `yaml:"address"`
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
	Rabbit      struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Port     int    `yaml:"port"`
		Address  string `yaml:"address"`
	} `yaml:"rabbit"`
	SelfCfg      `yaml:"mediaHost"`
	DownloadPath string `yaml:"-"`
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
	if err != nil {
		return
	}

	dlPath, err := getDownloadPath()
	if err != nil {
		return
	}

	s.DownloadPath = dlPath

	return
}
