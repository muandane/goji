package config

import (
	"testing"
)

func TestViperConfig(t *testing.T) {
	config, err := ViperConfig()
	if err != nil {
		t.Errorf("ViperConfig() error = %v", err)
	}
	if config == nil {
		t.Errorf("Expected a config, got nil")
	}

}
