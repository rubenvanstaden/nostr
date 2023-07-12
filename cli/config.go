package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rubenvanstaden/nostr"
)

type Config struct {
	Path       string          `json:"path"`
	PublicKey  string          `json:"publickey,omitempty"`
	PrivateKey string          `json:"privatekey,omitempty"`
	Profile    nostr.Profile   `json:"profile"`
	Relays     []string        `json:"relays,omitempty"`
	Following  map[string]User `json:"following,omitempty"`
}

type User struct {
	PublicKey string `json:"key"`
	Name      string `json:"name,omitempty"`
}

func DecodeConfig(path string) (*Config, error) {

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the file
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	if config.Path == "" {
		config.Path = path
	}

	return &config, nil
}

func (s *Config) Encode() {

	// Open the file
	file, err := os.OpenFile(s.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("error opening file:", err)
		return
	}
	defer file.Close()

	// Encode the new data
	encoder := json.NewEncoder(file)

	// Format: Pretty print to file.
	encoder.SetIndent("", "  ")

	// Write to file
	err = encoder.Encode(&s)
	if err != nil {
		fmt.Println("error encoding JSON:", err)
		return
	}

	log.Println("local config updated")
}
