package cmd

import (
	"fmt"

	"log"
	"os"
	"os/exec"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version     string
	versionFlag bool
	typeFlag    string
	scopeFlag   string
	messageFlag string
)

var rootCmd = &cobra.Command{
	Use:   "goji",
	Short: "Goji CLI",
	Long:  `Goji is a cli tool to generate conventional commits with emojis`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			color.Set(color.FgGreen)
			fmt.Printf("goji version: v%s\n", version)
			color.Unset()
			return
		}

		config, err := config.LoadConfig(".goji.json")
		if err != nil {
			log.Fatalf(color.YellowString("Error loading config file: %v"), err)
		}

		commitMessage, err := utils.AskQuestions(config)
		if err != nil {
			log.Fatalf(color.YellowString("Error asking questions: %v"), err)
		}

		gitCmd := exec.Command("git", "commit", "-m", commitMessage)
		output, err := gitCmd.CombinedOutput()
		if err != nil {
			fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
			fmt.Println("Git commit output: ", string(output))
			return
		}
		fmt.Printf("Git commit output: %s\n", string(output))
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Specify the type from the config file")
	rootCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Specify a custom scope")
	rootCmd.Flags().StringVarP(&messageFlag, "message", "m", "", "Specify a commit message")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
