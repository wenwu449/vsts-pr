package vsts

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	configPath = "VSTS_CONFIG_PATH"
)

type imageConfig struct {
	Os         string `json:"os"`
	ConfigPath string `json:"configPath"`
	Header     string `json:"header"`
}

type changeGroup []string

// Config is configuration for VSTS access
type Config struct {
	Username                 string        `json:"username"`
	Password                 string        `json:"password"`
	Instance                 string        `json:"instance"`
	Collection               string        `json:"collection"`
	Project                  string        `json:"project"`
	Repo                     string        `json:"repo"`
	MasterBranch             string        `json:"masterBranch"`
	UserID                   string        `json:"userId"`
	SupportLegacyImageFormat bool          `json:"supportLegacyImageFormat"`
	ImageConfigs             []imageConfig `json:"imageConfigs"`
	ChangeGroups             []changeGroup `json:"changeGroups"`
	StorageEntitiesPrefix    []string      `json:"storageEntitiesPrefix"`
	Endpoints                []string      `json:"endpoints"`
}

// GetConfig loads configuration from file
func GetConfig() (*Config, error) {
	configPathString := os.Getenv(configPath)
	if len(configPathString) == 0 {
		return nil, fmt.Errorf("env '%s' not found", configPath)
	}

	config := Config{}

	file, _ := os.Open(configPathString)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
