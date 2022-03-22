package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var config Config

// load config from 'config.json' file
func loadConfig(config *Config) {
	_config, _ := os.ReadFile("./setting/config.json")
	err := json.Unmarshal([]byte(_config), config)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Configuration has been read successfully!")
	}
}

// update config file with new config
func updateConfigFile() {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(config)

	err := os.WriteFile("./config.json", buf.Bytes(), os.ModeType)
	if err == nil {
		fmt.Println("Configuration has been updated successfully!")
	}
}
