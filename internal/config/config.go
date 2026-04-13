package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFile = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	fileBase, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return "", err
	}

	fileName := fileBase + "/" + configFile
	return fileName, nil
}

func Read() (Config, error) {

	fileName, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("Error getting config file name: %v\n", err)
		return Config{}, err
	}

	configData, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading confif file: %v\n", err)
		return Config{}, err
	}

	var currentConfig Config
	err = json.Unmarshal(configData, &currentConfig)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON data: %v\n", err)
		return Config{}, err
	}

	return currentConfig, nil
}

func (c *Config) SetUser(user string) error {

	c.CurrentUserName = user
	jsonData, err := json.Marshal(c)
	if err != nil {
		fmt.Printf("Failed to marshal configuration data: %v\n", err)
		return err
	}

	fileName, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("Error getting config file name: %v\n", err)
		return err
	}

	err = os.WriteFile(fileName, jsonData, 0666)
	if err != nil {
		fmt.Printf("Error writing to config file: %v\n", err)
		return err
	}

	return nil
}
