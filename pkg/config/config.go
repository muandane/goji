package config

import (
	"os/exec"

	"github.com/spf13/viper"
)

func ViperConfig() (*Config, error) {
	viper.SetConfigName(".goji")
	viper.SetConfigType("json")

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitDirBytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	gitDir := string(gitDirBytes)
	viper.AddConfigPath(gitDir)
	viper.AddConfigPath("$HOME")

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
