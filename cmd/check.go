/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the commit message follows the conventional commit format",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("sh", "-c", "git log -1 --pretty=%B").Output()
		if err != nil {
			log.Fatal(err)
		}

		commitMessage := strings.TrimSpace(string(out))

		re := regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore)\((.*)\): (.*)$`)
		if !re.MatchString(commitMessage) {
			fmt.Println("Error: Your commit message does not follow the conventional commit format.")
			os.Exit(1)
		} else {
			fmt.Println("Success: Your commit message follows the conventional commit format.")
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
