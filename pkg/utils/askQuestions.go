package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muandane/goji/pkg/config"
)

const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5A86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02AF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

type state int

const (
	statusNormal state = iota
	stateDone
)

// func AskQuestions(config *config.Config) (string, error) {
// 	var commitType string
// 	var commitScope string
// 	var commitSubject string
// 	commitTypeOptions := make([]huh.Option[string], len(config.Types))

// 	for i, ct := range config.Types {
// 		commitTypeOptions[i] = huh.NewOption[string](fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description), fmt.Sprintf("%s %s", ct.Name, ct.Emoji))
// 	}

// 	promptType := huh.NewSelect[string]().
// 		Title("Select the type of change you are committing:").
// 		Options(commitTypeOptions...).
// 		Value(&commitType)

// 	err := promptType.Run()
// 	if err != nil {
// 		return "", err
// 	}
// 	// Only ask for commitScope if not in SkipQuestions
// 	if !isInSkipQuestions("Scopes", config.SkipQuestions) {
// 		promptScope := huh.NewInput().
// 			Title("What is the scope of this change? (class or file name): (press [enter] to skip)").
// 			CharLimit(50).
// 			Placeholder("Example: ci, api, parser").
// 			Value(&commitScope)

// 		err = promptScope.Run()
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	promptSubject := huh.NewInput().
// 		Title("Write a short and imperative summary of the code changes: (lower case and no period)").
// 		CharLimit(100).
// 		Placeholder("Short description of your commit").
// 		Value(&commitSubject)

// 	err = promptSubject.Run()
// 	if err != nil {
// 		return "", err
// 	}

// 	var commitMessage string
// 	if commitScope == "" {
// 		commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
// 	} else {
// 		commitMessage = fmt.Sprintf("%s (%s): %s", commitType, commitScope, commitSubject)
// 	}
// 	return commitMessage, nil
// }

//	func isInSkipQuestions(value string, list []string) bool {
//		for _, v := range list {
//			if v == value {
//				return true
//			}
//		}
//		return false
//	}

type Model struct {
	config        *config.Config
	commitType    string
	commitScope   string
	commitSubject string
	form          *huh.Form
	width         int
	styles        *Styles
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.Copy().
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}
func NewModel(config *config.Config) Model {
	m := Model{
		config: config,
		width:  maxWidth,
		styles: NewStyles(lipgloss.DefaultRenderer()),
	}
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("commitType").
				Options(m.createCommitTypeOptions()...).
				Title("Select the type of change you are committing:").
				Value(&m.commitType),

			huh.NewInput().
				Key("commitScope").
				Title("What is the scope of this change? (class or file name): (press [enter] to skip)").
				CharLimit(50).
				Placeholder("Example: ci, api, parser").
				Value(&m.commitScope),

			huh.NewInput().
				Key("commitSubject").
				Title("Write a short and imperative summary of the code changes: (lower case and no period)").
				CharLimit(100).
				Placeholder("Short description of your commit").
				Value(&m.commitSubject),
		),
	).WithShowHelp(false).WithShowErrors(false)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	return m.form.View()
}

func (m Model) createCommitTypeOptions() []huh.Option[string] {
	commitTypeOptions := make([]huh.Option[string], len(m.config.Types))

	for i, ct := range m.config.Types {
		commitTypeOptions[i] = huh.NewOption[string](fmt.Sprintf("%-10s %-5s %-10s", ct.Name, ct.Emoji, ct.Description), fmt.Sprintf("%s %s", ct.Name, ct.Emoji))
	}

	return commitTypeOptions
}
