package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

const navigationInfoText = "(esc to go back to the previous step, ctrl+c to quit)"

type model struct {
	selectedFile    string
	namespace       string
	secretName      string
	secretKey       string
	filePicker      *filePicker
	namespaceInput  textinput.Model
	secretNameInput textinput.Model
	secretKeyInput  textinput.Model
	currentState    viewState
}

type viewState string

const (
	viewStateFilePicker viewState = "filepicker"
	viewStateNamespace  viewState = "namespace"
	viewStateSecretName viewState = "secretname"
	viewStateSecretKey  viewState = "secretkey"
	viewStateCert       viewState = "cert"
	viewStateResult     viewState = "result"
)

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			switch m.currentState {
			case viewStateFilePicker: // nothing to do here, it's the first view state
			case viewStateNamespace:
				m.currentState = viewStateFilePicker
			case viewStateSecretName:
				m.currentState = viewStateNamespace
			case viewStateSecretKey:
				m.currentState = viewStateSecretName
			}
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			switch m.currentState {
			case viewStateFilePicker:
				selection := m.filePicker.table.SelectedRow()[0]
				path := filepath.Join(m.filePicker.currentDir, selection)

				if strings.HasSuffix(selection, dirAffix) {
					return m, m.filePicker.navigateToCurrentSelection(msg, selection)
				}

				m.selectedFile = path

				m.currentState = viewStateNamespace
				m.namespaceInput.Blink()
				m.namespaceInput.Focus()
			case viewStateNamespace:
				m.namespace = m.namespaceInput.Value()

				m.currentState = viewStateSecretName
				m.secretNameInput.Blink()
				m.secretNameInput.Focus()
			case viewStateSecretName:
				m.secretName = m.secretNameInput.Value()

				m.currentState = viewStateSecretKey
				m.secretKeyInput.Blink()
				m.secretKeyInput.Focus()
			case viewStateSecretKey:
				m.secretKey = m.secretKeyInput.Value()

				// TODO seal and create file
			}
		}
	}

	switch m.currentState {
	case viewStateFilePicker:
		cmd = m.filePicker.Update(msg)
	case viewStateNamespace:
		m.namespaceInput, cmd = m.namespaceInput.Update(msg)
	case viewStateSecretName:
		m.secretNameInput, cmd = m.secretNameInput.Update(msg)
	case viewStateSecretKey:
		m.secretKeyInput, cmd = m.secretKeyInput.Update(msg)
	}

	return m, cmd
}

func (m *model) View() string {
	switch m.currentState {
	case viewStateFilePicker:
		return fmt.Sprintf("%s\n\n%s\n",
			baseStyle.Render(m.filePicker.table.View()),
			navigationInfoText, // TODO update text for first step
		)
	case viewStateNamespace:
		return fmt.Sprintf(
			"Selected file: %s\n\nWhat is your namspace?\n\n%s\n\n%s\n",
			m.selectedFile,
			m.namespaceInput.View(),
			navigationInfoText,
		)
	case viewStateSecretName:
		return fmt.Sprintf(
			"Choose a secret name:\n\n%s\n\n%s\n",
			m.secretNameInput.View(),
			navigationInfoText,
		)
	case viewStateSecretKey:
		return fmt.Sprintf(
			"Choose a secret key:\n\n%s\n\n%s\n",
			m.secretKeyInput.View(),
			navigationInfoText,
		)
	default:
		return "something went wrong"
	}
}

func StartUI(c *cli.Context) error {
	// TODO maybe use one view for all text-inputs
	// TODO add cert selection
	// TODO move kubectl logic from goseal.go to kubectl.go
	// TODO fix filepicker navigation

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if err := tea.NewProgram(
		&model{
			filePicker:      initFilePicker(currentDir),
			currentState:    viewStateFilePicker,
			namespaceInput:  initTextInput("my-kubernetes-namespace"),
			secretNameInput: initTextInput("my-secure-secret"),
			secretKeyInput:  initTextInput("my-file.yaml"),
		},
		//tea.WithAltScreen(),
	).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return nil
}

func initTextInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 156
	input.Width = 20

	return input
}
