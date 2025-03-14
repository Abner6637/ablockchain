package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type GenesisConfig struct {
	Difficulty uint64            `json:"difficulty"`
	ChainID    uint64            `json:"chainId"`
	Alloc      map[string]uint64 `json:"alloc"`
}

func LoadGenesisConfig(path string) (*GenesisConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open genesis file: %v", err)
	}
	defer file.Close()

	var config GenesisConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse genesis file: %v", err)
	}
	return &config, nil

}
