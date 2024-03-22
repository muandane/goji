package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func ViperConfig() (*Config, error) {
	viper.SetConfigName(".goji")
	viper.SetConfigType("json")

	gitDir, err := GetGitRootDir()
	if err != nil {
		return nil, err
	}
	homeDir, _ := os.UserHomeDir()
	_, err = os.Stat(filepath.Join(gitDir, ".goji.json"))
	if err == nil {
		viper.AddConfigPath(gitDir)
	} else if os.IsNotExist(err) {
		viper.AddConfigPath(homeDir)
	} else {
		return nil, err
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("unable to find .goji.json in %s or %s", gitDir, homeDir)
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
