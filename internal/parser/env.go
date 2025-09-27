package parser

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type AuthConfig struct {
	Type         string `json:"Type"`
	TokenURL     string `json:"Token Url"`
	GrantType    string `json:"Grant Type"`
	ClientID     string `json:"Client ID"`
	ClientSecret string `json:"Client Secret"`
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	Scope        string `json:"Scope"`
}

type AuthMap map[string]AuthConfig

type Security struct {
	Auth AuthMap `json:"Auth"`
}

type Environment struct {
	Security Security `json:"Security"`
}

type Config map[string]Environment

func ParseEnv(env string) (Environment, error) {
	parts := strings.Split(env, ":")
	if len(parts) != 2 {
		return Environment{}, errors.New("invalid environment format")
	}
	fileName := parts[0]
	envName := parts[1]

	file, err := os.ReadFile(fileName)
	if err != nil {
		return Environment{}, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return Environment{}, errors.New("invalid environment content")
	}

	return cfg[envName], nil
}
