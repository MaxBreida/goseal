package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

type viewModel struct{}

func (v viewModel) View() string {
	return "Hello World"
}

func (v viewModel) Init() tea.Cmd {
	return nil
}

func (v viewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return v, tea.Quit
		}
	}

	return v, nil
}

func startUI(c *cli.Context) error {
	ui := tea.NewProgram(viewModel{})

	if err := ui.Start(); err != nil {
		return err
	}

	os.Exit(0)

	return nil
}
