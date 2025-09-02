package app

import (
	"flag"
	"os"
	"path"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func GetBasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".mediapire", "mediahost"), nil
}

func getDownloadPath() (string, error) {
	basePath, err := GetBasePath()
	if err != nil {
		return "", err
	}

	return path.Join(basePath, "downloads"), nil
}

func getArtPath() (string, error) {
	basePath, err := GetBasePath()
	if err != nil {
		return "", err
	}

	return path.Join(basePath, "art"), nil
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
	Name        string    `yaml:"name"`
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
	ArtPath      string `yaml:"-"`
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

	if s.Name == "" {
		log.Error().Msg("Must provide a name in the config file")
		os.Exit(1)
	}

	dlPath, err := getDownloadPath()
	if err != nil {
		return
	}

	s.DownloadPath = dlPath

	artPath, err := getArtPath()
	if err != nil {
		return
	}

	s.ArtPath = artPath

	return
}
