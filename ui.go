package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

const navigationInfoText = "(esc to go back to the previous step, ctrl+c to quit)"

type model struct {
	// secretFile
	selectedFile     string
	secretFilePicker *filePicker

	// kubectl specific values
	focusIndex int
	textInputs []textinput.Model

	// certFile
	certFile       string
	certFilePicker *filePicker

	// fileMode
	fileModeCursor int
	fileMode       string

	currentState viewState
	prevState    viewState
}

type viewState string

const (
	viewStateFilePicker viewState = "filepicker"
	viewStateFileMode   viewState = "filemode"
	viewStateTextInputs viewState = "textinputs"
	viewStateCert       viewState = "cert"
	viewStateResult     viewState = "result"
)

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		// go back to previous view state
		case "esc":
			switch m.currentState {
			case viewStateFilePicker:
				return m, tea.Quit
			case viewStateFileMode:
				m.currentState = viewStateFilePicker
			case viewStateTextInputs:
				if m.prevState == viewStateFileMode {
					m.currentState = viewStateFileMode
					break
				}
				m.currentState = viewStateFilePicker
			case viewStateCert:
				m.currentState = viewStateTextInputs
			}

		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
			switch m.currentState {
			case viewStateTextInputs:
				s := msg.String()

				// Cycle indexes
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.textInputs) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.textInputs)
				}

				cmds := m.setFocusedTextInput()

				return m, tea.Batch(cmds...)
			case viewStateFileMode:
				s := msg.String()

				// Cycle indexes
				if s == "up" || s == "shift+tab" {
					m.fileModeCursor--
				} else {
					m.fileModeCursor++
				}

				if m.fileModeCursor >= len(fileModes) {
					m.fileModeCursor = 0
				} else if m.fileModeCursor < 0 {
					m.fileModeCursor = len(fileModes) - 1
				}
			default:
				break
			}

		// enter selection
		case "enter":
			switch m.currentState {
			case viewStateFilePicker:
				selection := m.secretFilePicker.table.SelectedRow()[0]
				path := filepath.Join(m.secretFilePicker.filePath, selection)

				if strings.HasSuffix(selection, dirAffix) {
					return m, m.secretFilePicker.navigateToCurrentSelection(msg, path)
				}

				m.selectedFile = path

				if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
					m.currentState = viewStateFileMode

					break
				}

				m.fileMode = fileModes[0] // if we don't have a yaml file, only file is eligable for now
				m.currentState = viewStateTextInputs
				m.textInputs[0].Focus()
			case viewStateFileMode:
				m.fileMode = fileModes[m.fileModeCursor]

				m.prevState = m.currentState
				m.currentState = viewStateTextInputs
				m.textInputs[0].Focus()
			case viewStateTextInputs:
				// Did the user press enter while the submit button was focused?
				// If so, exit.
				for i := range m.textInputs {
					if m.textInputs[i].Value() == "" {
						m.focusIndex = i

						cmds := m.setFocusedTextInput()

						return m, tea.Batch(cmds...)
					}
				}

				if m.focusIndex == len(m.textInputs) {
					m.currentState = viewStateCert
				}
			case viewStateCert:
				selection := m.certFilePicker.table.SelectedRow()[0]
				path := filepath.Join(m.certFilePicker.filePath, selection)

				if strings.HasSuffix(selection, dirAffix) {
					return m, m.certFilePicker.navigateToCurrentSelection(msg, path)
				}

				m.certFile = path
				m.currentState = viewStateResult

				secrets, err := GetSecretsFromFile(m.fileMode, m.selectedFile, m.textInputs[2].Value())
				if err != nil {
					panic(err)
				}

				sealedSecret, err := SealSecret(secrets, m.textInputs[1].Value(), m.textInputs[0].Value(), m.certFile)
				if err != nil {
					panic(err)
				}

				err = os.WriteFile("secret.yaml", sealedSecret, 0644)
				if err != nil {
					panic(err)
				}
			case viewStateResult:
				return m, tea.Quit
			}
		}
	}

	switch m.currentState {
	case viewStateFilePicker:
		cmd = m.secretFilePicker.Update(msg)
	case viewStateTextInputs:
		cmd = m.updateInputs(msg)
	case viewStateCert:
		cmd = m.certFilePicker.Update(msg)
	}

	return m, cmd
}

const (
	fileModeFile = "file"
	fileModeYaml = "yaml"
)

var fileModes = []string{fileModeFile, fileModeYaml}

func (m *model) View() string {
	switch m.currentState {
	case viewStateFilePicker:
		return fmt.Sprintf("%s\n\n%s\n",
			baseStyle.Render(m.secretFilePicker.table.View()),
			navigationInfoText, // TODO update text for first step
		)
	case viewStateFileMode:
		b := strings.Builder{}
		b.WriteString("Which file mode should be used for this secret?\n\n")

		for i := 0; i < len(fileModes); i++ {
			if m.fileModeCursor == i {
				b.WriteString("(•) ")
			} else {
				b.WriteString("( ) ")
			}
			b.WriteString(fileModes[i])
			b.WriteString("\n")
		}
		b.WriteString(navigationInfoText)

		return b.String()
	case viewStateTextInputs:
		var b strings.Builder

		fmt.Fprintf(&b, "\nSelected file: %s\n", m.selectedFile)
		fmt.Fprintf(&b, "Mode: %s\n\n", m.fileMode)

		for i := range m.textInputs {
			b.WriteString(m.textInputs[i].View())
			if i < len(m.textInputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.textInputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)

		b.WriteString(navigationInfoText)

		return b.String()
	case viewStateCert:
		return fmt.Sprintf("%s\n\n%s\n",
			baseStyle.Render(m.certFilePicker.table.View()),
			navigationInfoText,
		)
	case viewStateResult:
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		return fmt.Sprintf("Secret sucessfully sealed. Filepath: %s/secret.yaml\n\n%s\n",
			wd,
			"Press Enter or Ctrl+C to quit.",
		)
	}

	return "invalid view state"
}

// StartUI starts the TUI.
func StartUI(c *cli.Context) error {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	if err := tea.NewProgram(
		&model{
			currentState:     viewStateFilePicker,
			secretFilePicker: initFilePicker(currentDir),
			textInputs:       initTextInputs(),
			certFilePicker:   initFilePicker(homeDir),
		},
		tea.WithAltScreen(),
	).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return nil
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Render("[ Continue ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Continue"))
)

func initTextInputs() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	for i := range inputs {
		var placeholder string

		switch i {
		case 0:
			placeholder = "Namespace"
		case 1:
			placeholder = "Secret Name"
		case 2:
			placeholder = "Secret Key"
		}

		inputs[i] = initTextInput(placeholder)
	}

	return inputs
}

func initTextInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CursorStyle = cursorStyle
	input.CharLimit = 32

	return input
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *model) setFocusedTextInput() []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	for i := 0; i <= len(m.textInputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.textInputs[i].Focus()
			m.textInputs[i].PromptStyle = focusedStyle
			m.textInputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.textInputs[i].Blur()
		m.textInputs[i].PromptStyle = noStyle
		m.textInputs[i].TextStyle = noStyle
	}

	return cmds
}
