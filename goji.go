package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

var helpFlag bool
var versionFlag bool

type CommitType struct {
	Emoji       string
	Code        string
	Description string
	Name        string
}

type Config struct {
	Types            []CommitType
	Scopes           []string
	Symbol           bool
	SkipQuestions    []string
	Questions        map[string]string
	SubjectMaxLength int
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func AskQuestions(config *Config) (string, error) {
	var commitType string
	var commitScope string
	var commitSubject string

	commitTypeOptions := make([]string, len(config.Types))
	for i, ct := range config.Types {
		commitTypeOptions[i] = fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description)
	}

	promptType := &survey.Select{
		Message: "Select the type of change you are committing:",
		Options: commitTypeOptions,
	}
	err := survey.AskOne(promptType, &commitType)
	if err != nil {
		return "", err
	}

	promptScope := &survey.Input{
		Message: "Enter the scope of the change:",
	}
	err = survey.AskOne(promptScope, &commitScope)
	if err != nil {
		return "", err
	}

	promptSubject := &survey.Input{
		Message: "Enter a short description of the change:",
	}
	err = survey.AskOne(promptSubject, &commitSubject)
	if err != nil {
		return "", err
	}

	commitMessage := fmt.Sprintf("%s (%s): %s", commitType, commitScope, commitSubject)
	return commitMessage, nil
}
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
	fmt.Println("Goji v", version, "is a cli tool to generate conventional commits with emojis.", "\n")
	color.Unset()
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf(color.YellowString("Error loading config: %v"), err)
	}

	commitMessage, err := AskQuestions(config)
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
