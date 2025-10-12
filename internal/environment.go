package internal

import (
	"encoding/json"
	"fmt"
)

// Config maps environment names to their respective configurations.
// It serves as the top-level structure for managing multiple environments
// and their variable and authentication definition.
type Config map[string]Environment

// Environment defines a named configuration context that can include variables
// and authentication settings. It’s typically used to separate configurations
// for different stages like development, staging, or production.
type Environment struct {
	Variables map[string]string
	Security  Security `json:"Security"`
}

type Security struct {
	Auth AuthMap `json:"Auth"`
}

type AuthMap map[string]AuthConfig

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

func (e *Environment) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if secRaw, ok := raw["Security"]; ok {
		if err := json.Unmarshal(secRaw, &e.Security); err != nil {
			return err
		}
		delete(raw, "Security")
	}

	e.Variables = make(map[string]string, len(raw))
	for k, v := range raw {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("key %s is not a string: %w", k, err)
		}
		e.Variables[k] = s
	}

	return nil
}
