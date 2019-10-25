package config

import (
	"encoding/json"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg := Load("./")
	bz, _ := json.Marshal(cfg)
	println(string(bz))
}
