package config

import "github.com/muandane/goji/pkg/models"

type Config struct {
	Types             []models.CommitType // Keep this as your primary commit types
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
	AIProvider        string    `mapstructure:"aiProvider"`
	AIChoices         AIChoices `mapstructure:"aiChoices"`
}

type Gitmoji struct {
	Emoji       string `json:"emoji"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type initConfig struct {
	Types            []Gitmoji `json:"types"` // Keep this for initialization
	Scopes           []string  `json:"scopes"`
	SkipQuestions    []string  `json:"skipQuestions"`
	SubjectMaxLength int       `json:"subjectMaxLength"`
	SignOff          bool      `json:"signOff"`
	NoEmoji          bool      `json:"noemoji"`
	AIProvider       string    `json:"aiProvider"`
	AIChoices        AIChoices `json:"aiChoices"`
}

// AIChoices and AIConfig remain the same
type AIChoices struct {
	OpenAI     AIConfig `mapstructure:"openai" json:"openai"`
	Groq       AIConfig `mapstructure:"groq" json:"groq"`
	Claude     AIConfig `mapstructure:"claude" json:"claude"`
	Ollama     AIConfig `mapstructure:"ollama" json:"ollama"`
	OpenRouter AIConfig `mapstructure:"openrouter" json:"openrouter"`
	Deepseek   AIConfig `mapstructure:"deepseek" json:"deepseek"`
	Gemini     AIConfig `mapstructure:"gemini" json:"gemini"`
}

type AIConfig struct {
	Model string `mapstructure:"model" json:"model"`
}
