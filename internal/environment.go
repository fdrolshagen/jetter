package internal

import (
	"encoding/json"
	"fmt"
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
	Variables map[string]string
	Security  Security `json:"Security"`
}

type Config map[string]Environment

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
