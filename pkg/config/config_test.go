package config

import (
	"testing"
)

func TestViperConfig(t *testing.T) {
	_, err := ViperConfig()
	if err != nil {
		t.Errorf("ViperConfig() error = %v", err)
	}
}
