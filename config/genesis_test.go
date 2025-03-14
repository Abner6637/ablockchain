package config

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLoadGenesisConfig(t *testing.T) {
	config, err := LoadGenesisConfig("./genesis.json")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Error marshalling GenesisConfig:%v", err)
	}
	// 将解析出的config转换为json格式打印出来
	fmt.Println(string(configJSON))
}
