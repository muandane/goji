package main

import (
	"flag"
	"fmt"
	"goji/pkg/config"
	"goji/pkg/utils"
	"log"
	"os/exec"

	"github.com/fatih/color"
)

var helpFlag bool
var versionFlag bool

func init() {
	flag.BoolVar(&helpFlag, "h", false, "Display help information")
	flag.BoolVar(&helpFlag, "help", false, "Display help information")
	flag.BoolVar(&versionFlag, "v", false, "Display version information")
	flag.BoolVar(&versionFlag, "version", false, "Display version information")
}

func main() {
	version := "0.0.2"
	flag.Parse()
	if helpFlag {
		fmt.Println("Help information:")
		fmt.Println("-h --help: Display help information")
		fmt.Println("-v --version: Display version information")
		return
	}

	if versionFlag {
		fmt.Println("CLI version: ", version)
		return
	}
	color.Set(color.FgGreen)
	fmt.Println("Goji v", version, "is a cli tool to generate conventional commits with emojis.")
	color.Unset()

	config, err := config.LoadConfig(".goji.json")
	if err != nil {
		log.Fatalf(color.YellowString("Error loading config: %v"), err)
	}

	commitMessage, err := utils.AskQuestions(config)
	if err != nil {
		log.Fatalf(color.YellowString("Error asking questions: %v"), err)
	}
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
		return
	}
	fmt.Printf("Git commit output: %s\n", string(output))
}
