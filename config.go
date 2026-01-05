package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogDir string `yaml:"logDir"`
}

func LoadConfig() *Config {
	// Default configuration
	cfg := &Config{
		LogDir: os.TempDir(),
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
	if !filepath.IsAbs(cfg.LogDir) {
		exeDir := filepath.Dir(exePath)
		cfg.LogDir = filepath.Join(exeDir, cfg.LogDir)
	}

	return cfg
}
