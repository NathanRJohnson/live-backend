package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	ServerPort  uint16
	SecretsPath string
	Secrets     FirebaseSecrets
}

func LoadConfig() Config {
	currentDir, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		fmt.Println("Error getting current directory", err)
	}

	// local config init
	cfg := Config{
		ServerPort:  3000,
		SecretsPath: filepath.Join(currentDir, "../secrets/firebase-serviceKey.json"),
	}

	// load variables from docker environment
	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	if _, exists := os.LookupEnv("SECRETS_PATH"); exists {
		cfg.SecretsPath = "/run/secrets/serviceKey"
	}

	secrets, err := loadSecrets(cfg.SecretsPath)
	if err != nil {
		fmt.Println("Cannot load secrets", err)
	}
	cfg.Secrets = *secrets

	return cfg
}

type FirebaseSecrets struct {
	ProjectID string `json:"project_id"`
}

func loadSecrets(secretsPath string) (*FirebaseSecrets, error) {
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		fmt.Println("error reading secrets file:", err)
		return nil, err
	}

	var secrets FirebaseSecrets
	err = json.Unmarshal(data, &secrets)
	if err != nil {
		fmt.Println("error parsing secrets", err)
		return nil, err
	}

	return &secrets, nil
}
