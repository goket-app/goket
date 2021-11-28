package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/wojciechka/goket/pkg/eventprocessor"
)

// Config is JSON configuration read by the client
type Config struct {
	Keys eventprocessor.EventMap `json:"keys"`
}

func getConfig(configPath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
