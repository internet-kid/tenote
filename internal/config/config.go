package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configPerm = 0o644

// AppConfig holds user-configurable settings persisted to disk.
type AppConfig struct {
	StorageDir string `json:"storage_dir"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "tenote", "config.json"), nil
}

func defaultStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".local", "share", "tenote"), nil
}

// LoadConfig reads the config file; returns defaults if the file does not exist.
func LoadConfig() (AppConfig, error) {
	path, err := configFilePath()
	if err != nil {
		return AppConfig{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig()
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.StorageDir == "" {
		return defaultConfig()
	}
	return cfg, nil
}

// SaveConfig writes the config to disk, creating the config directory if needed.
func SaveConfig(cfg AppConfig) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, configPerm); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func defaultConfig() (AppConfig, error) {
	dir, err := defaultStorageDir()
	if err != nil {
		return AppConfig{}, err
	}
	return AppConfig{StorageDir: dir}, nil
}
