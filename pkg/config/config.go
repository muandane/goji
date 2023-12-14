package config

import (
	"fmt"
	"os/exec"

	"github.com/spf13/viper"
)

func ViperConfig() (*Config, error) {
	viper.SetConfigName(".goji")
	viper.SetConfigType("json")

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitDirBytes, err := cmd.Output()
	if err != nil {
		_ = fmt.Errorf("error finding git root directory: %v", err)
		return nil, err
	}
	gitDir := string(gitDirBytes)
	viper.AddConfigPath(gitDir)
	viper.AddConfigPath("$HOME")

	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("unable to decode into struct, %v", err)
	}
	return &config, nil
}
