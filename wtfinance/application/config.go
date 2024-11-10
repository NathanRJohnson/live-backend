package application

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	ServerPort  uint16
	SecretsPath string
	ServiceKey  []byte
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
		cfg.SecretsPath = "/run/secrets/googleSheets"
	}

	secrets, err := os.ReadFile(cfg.SecretsPath)
	if err != nil {
		log.Fatalf("Unable to read secrets: %v", err)
	}

	cfg.ServiceKey = secrets

	return cfg
}
