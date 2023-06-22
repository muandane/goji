package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"goji/pkg/config"
	"goji/pkg/utils"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func init() {
}

func main() {
	version := "0.0.1"
	helpFlag := flag.Bool("help", false, "Display help information")
	versionFlag := flag.Bool("version", false, "Display version information")
	initFlag := flag.Bool("init", false, "Generate a configuration file")

	flag.Parse()
	if *helpFlag {
		fmt.Println("Help information:")
		fmt.Println("-h --help: Display help information")
		fmt.Println("-v --version: Display version information")
		fmt.Println("-i --init: Generate a configuration file")
		return
	}

	if *versionFlag {
		fmt.Println("CLI version: ", version)
		return
	}

	color.Set(color.FgGreen)
	fmt.Println("Goji v", version, " is a cli tool to generate conventional commits with emojis.")
	color.Unset()

	if *initFlag {
		gitmojis := AddCustomCommitTypes([]Gitmoji{})
		config := initConfig{
			Types:            gitmojis,
			Scopes:           []string{"home", "accounts", "ci"},
			Symbol:           true,
			SkipQuestions:    []string{},
			SubjectMaxLength: 50,
		}
		err := SaveGitmojisToFile(config, ".goji.json")
		if err != nil {
			fmt.Printf("Error saving gitmojis to file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Gitmojis saved to .goji.json 🎊")
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
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
		return
	}
	fmt.Printf("Git commit output: %s\n", string(output))
}

type Gitmoji struct {
	Emoji       string `json:"emoji"`
	Entity      string `json:"entity"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Name        string `json:"name"`
}
type initConfig struct {
	Types            []Gitmoji `json:"Types"`
	Scopes           []string  `json:"Scopes"`
	Symbol           bool      `json:"Symbol"`
	SkipQuestions    []string  `json:"SkipQuestions"`
	SubjectMaxLength int       `json:"SubjectMaxLength"`
}

func AddCustomCommitTypes(gitmojis []Gitmoji) []Gitmoji {
	customGitmojis := []Gitmoji{
		{Emoji: "✨", Code: ":sparkles:", Description: "Introduce new features.", Name: "feature"},
		{Emoji: "🐛", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
		{Emoji: "🚧", Code: ":construction:", Description: "Work in progress.", Name: "wip"},
		{Emoji: "📦", Code: ":package:", Description: "Add or update compiled files or packages.", Name: "package"},
	}

	return append(gitmojis, customGitmojis...)
}

func GetGitRootDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitDirBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error finding git root directory: %v", err)
	}
	gitDir := string(gitDirBytes)
	gitDir = strings.TrimSpace(gitDir) // Remove newline character at the end

	return gitDir, nil
}

func SaveGitmojisToFile(config initConfig, filename string) error {
	gitDir, err := GetGitRootDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(gitDir, filename)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configFile, data, 0644)
}
