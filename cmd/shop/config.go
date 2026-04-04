package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	configFileName = ".rpgshopconfig.json"
)

type Config struct {
	CurrentUserID   uuid.UUID `json:"current_user_id"`
	CurrentUserName string    `json:"current_user_name"`
}

func (c *Config) SetUser(userID uuid.UUID, userName string) error {
	c.CurrentUserID = userID
	c.CurrentUserName = userName
	return write(*c)
}

func write(cfg Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	path, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error getting config path: %v", err)
	}

	return os.WriteFile(path, data, 0o644)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	return filepath.Join(home, configFileName), nil
}

func Read() (Config, error) {
	// get file path
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("error getting config file path: %v", err)
	}

	// open and read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}

	// unmarshal into struct
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return cfg, nil
}
