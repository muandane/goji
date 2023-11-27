/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/muandane/goji/pkg/config"
	"github.com/spf13/cobra"
)

var (
	globalFlag bool
	repoFlag   bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your goji config",
	Long:  `Initialize your goji config in your repository or globally in your home directory`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Set(color.FgYellow)
		err := config.InitRepoConfig(globalFlag, repoFlag)
		if err != nil {
			fmt.Printf("Failed to initialize repository configuration: %v\n", err)
			return
		}
		color.Unset()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&globalFlag, "global", false, "save the init file to your home directory")
	initCmd.Flags().BoolVar(&repoFlag, "repo", false, "save the init file in the repository")
}
