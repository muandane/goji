package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"goji/pkg/config"
	"goji/pkg/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func init() {
}

var (
	version = ""
)

func main() {

  
	helpFlag := flag.Bool("h", false, "Display help information")
	flag.BoolVar(helpFlag, "help", false, "display help")
	versionFlag := flag.Bool("v", false, "Display version information")
	flag.BoolVar(versionFlag, "version", false, "display help")
	initFlag := flag.Bool("i", false, "Generate a configuration file")
	flag.BoolVar(initFlag, "init", false, "display help")
	typeFlag := flag.String("t", "", "Specify the type from the config file")
	scopeFlag := flag.String("s", "", "Specify a custom scope")
	messageFlag := flag.String("m", "", "Specify a commit message")
	flag.StringVar(typeFlag, "type", "", "Specify the type from the config file")
	flag.StringVar(scopeFlag, "scope", "", "Specify a custom scope")
	flag.StringVar(messageFlag, "message", "", "Specify a commit message")

	flag.Parse()
	if *helpFlag {
		color.Set(color.FgGreen)
		fmt.Printf("goji v%s is a cli tool to generate conventional commits with emojis.\n", version)
		color.Unset()
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println(" goji")
		fmt.Println()
		fmt.Println("Help information:")
		fmt.Println(" -h --help: Display help information")
		fmt.Println(" -v --version: Display version information")
		fmt.Println(" -i --init: Generate a configuration file")
		return
	}

	if *versionFlag {
		color.Set(color.FgGreen)
		fmt.Printf("goji version: v%s\n", version)
		color.Unset()
		return
	}

	color.Set(color.FgGreen)
	fmt.Printf("Goji v%s is a cli tool to generate conventional commits with emojis.\n", version)
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

	if *typeFlag != "" && *messageFlag != "" {
		config, err := config.LoadConfig(".goji.json")
		if err != nil {
			log.Fatalf(color.YellowString("Error loading config file: %v"), err)
		}
		var typeMatch string
		for _, t := range config.Types {
			if *typeFlag == t.Name {
				typeMatch = fmt.Sprintf("%s %s", t.Name, t.Emoji)
				break
			}
		}

		// If no match was found, use the type flag as is
		if typeMatch == "" {
			typeMatch = *typeFlag
		}

		// Construct the commit message from the flags
		commitMessage := *messageFlag
		if typeMatch != "" {
			commitMessage = fmt.Sprintf("%s: %s", typeMatch, commitMessage)
			if *scopeFlag != "" {
				commitMessage = fmt.Sprintf("%s(%s): %s", typeMatch, *scopeFlag, *messageFlag)
			}
		}

		// Create the git commit command with the commit message
		cmd := exec.Command("git", "commit", "-m", commitMessage)

		// Run the command and capture the output and any error
		output, err := cmd.CombinedOutput()

		// Check if the command resulted in an error
		if err != nil {
			// If there was an error, print it and the output from the command
			fmt.Printf(color.MagentaString("Error executing git commit: %v\n"), err)
			fmt.Println("Git commit output: ", string(output))
			return
		}

		// If there was no error, print the output from the command
		fmt.Printf("Git commit output: %s\n", string(output))
	} else {
		// If not all flags were provided, fall back to your existing interactive prompt logic
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
			fmt.Println("Git commit output: ", string(output))
			return
		}
		fmt.Printf("Git commit output: %s\n", string(output))
	}
}

type Gitmoji struct {
	Emoji       string `json:"emoji"`
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
		{Emoji: "✨", Code: ":sparkles:", Description: "Introduce new features.", Name: "feat"},
		{Emoji: "🐛", Code: ":bug:", Description: "Fix a bug.", Name: "fix"},
		{Emoji: "📚", Code: ":books:", Description: "Documentation change.", Name: "docs"},
		{Emoji: "🎨", Code: ":art:", Description: "Improve structure/format of the code.", Name: "refactor"},
		{Emoji: "🧹", Code: ":broom:", Description: "A chore change.", Name: "chore"},
		{Emoji: "🧪", Code: ":test_tube:", Description: "Add a test.", Name: "test"},
		{Emoji: "🚑️", Code: ":ambulance:", Description: "Critical hotfix.", Name: "hotfix"},
		{Emoji: "⚰️", Code: ":coffin:", Description: "Remove dead code.", Name: "deprecate"},
		{Emoji: "⚡️", Code: ":zap:", Description: "Improve performance.", Name: "perf"},
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

	return os.WriteFile(configFile, data, 0644)
}