package main

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

type Command struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Command      string `json:"command"`
	SupportsArgs bool   `json:"supports_args"`
}

type Config struct {
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Commands    []Command `json:"commands"`
}

func getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".cmdcenter", "config.json"), nil
}

func loadConfigFile() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfigFile(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, content, 0644)
}
