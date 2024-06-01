package config

import "github.com/muandane/goji/pkg/models"

type Config struct {
	Types             []models.CommitType
	Scopes            []string
	SkipQuestions     []string
	Questions         map[string]string
	SubjectMaxLength  int
	SignOff           bool
	NoEmoji           bool
	CommitType        string
	CommitScope       string
	CommitSubject     string
	CommitDescription string
}

type Gitmoji struct {
	Emoji       string `json:"emoji"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Name        string `json:"name"`
}
type initConfig struct {
	Types            []Gitmoji `json:"types"`
	Scopes           []string  `json:"scopes"`
	SkipQuestions    []string  `json:"skipQuestions"`
	SubjectMaxLength int       `json:"subjectMaxLength"`
	SignOff          bool      `json:"signOff"`
	NoEmoji          bool      `json:"noEmoji"`
}
