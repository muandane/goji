package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
)

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
		commitTypeOptions[i] = fmt.Sprintf("%s %s", ct.Emoji, ct.Name)
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

func main() {
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	commitMessage, err := AskQuestions(config)
	if err != nil {
		log.Fatalf("Error asking questions: %v", err)
	}
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing git commit: %v\n", err)
		return
	}
	fmt.Printf("Git commit output: %s\n", string(output))
	// fmt.Println("Commit message:", commitMessage)
}
