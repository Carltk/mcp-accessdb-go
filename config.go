package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogDir string `yaml:"logDir"`
	TmpDir string `yaml:"tmpDir"`
	Debug  bool   `yaml:"debug"`
}

func LoadConfig() *Config {
	// Default configuration
	cfg := &Config{
		LogDir: "./log",
		TmpDir: os.TempDir(),
		Debug:  false,
	}

	exePath, err := os.Executable()
	if err != nil {
		return cfg
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg // Return defaults if file doesn't exist
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Printf("Error parsing config.yaml: %v. Using defaults.", err)
	}

	// Resolve relative paths based on the executable directory
	exePath, err := os.Executable()
	if err != nil {
		return cfg
	}
	exeDir := filepath.Dir(exePath)

	if !filepath.IsAbs(cfg.LogDir) {
		cfg.LogDir = filepath.Clean(filepath.Join(exeDir, cfg.LogDir))
	}
	if !filepath.IsAbs(cfg.TmpDir) {
		cfg.TmpDir = filepath.Clean(filepath.Join(exeDir, cfg.TmpDir))
	}

	return cfg
}
