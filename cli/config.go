package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rubenvanstaden/nostr"
)

type Config struct {
	Path       string            `json:"path"`
	PublicKey  string            `json:"publickey,omitempty"`
	PrivateKey string            `json:"privatekey,omitempty"`
	Profile    nostr.Profile     `json:"profile"`
	Relays     []string          `json:"relays,omitempty"`
	Following  map[string]Author `json:"following,omitempty"`
}

type Author struct {
	PublicKey string `json:"key"`
	Name      string `json:"name,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		Path:       "",
		PublicKey:  "",
		PrivateKey: "",
		Profile:    nostr.Profile{},
		Relays:     []string{},
		Following:  make(map[string]Author),
	}
}

func DecodeConfig(path string) (*Config, error) {

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the file
	config := NewConfig()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	if config.Path == "" {
		config.Path = path
	}

	return config, nil
}

// Save change to inmem data structure to persistent local file.
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

	log.Println("[-] Config file updated")
}
