package parser

import (
	"encoding/json"
	"errors"
	"github.com/fdrolshagen/jetter/internal"
	"os"
	"strings"
)

func ParseEnv(env string) (internal.Environment, error) {
	parts := strings.Split(env, ":")
	if len(parts) != 2 {
		return internal.Environment{}, errors.New("invalid environment format")
	}
	fileName := parts[0]
	envName := parts[1]

	file, err := os.ReadFile(fileName)
	if err != nil {
		return internal.Environment{}, err
	}

	var cfg internal.Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return internal.Environment{}, errors.New("invalid environment content")
	}

	envConfig, ok := cfg[envName]
	if !ok {
		return internal.Environment{}, errors.New("environment not found: " + envName)
	}

	return envConfig, nil
}
