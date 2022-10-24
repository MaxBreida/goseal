package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	currentDir string
	table      table.Model
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selection := m.table.SelectedRow()[0]

			if strings.HasSuffix(selection, "/") {
				m.currentDir = filepath.Join(m.currentDir, selection)
				m.table.SetRows(getTableRowsFromDir(m.currentDir))
				m.table.GotoTop()
			} else {
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", filepath.Join(m.currentDir, selection)),
				)
			}

			// TODO pass file to next step
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Size (B)", Width: 10},
		{Title: "Date modified", Width: 20},
	}

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(getTableRowsFromDir(currentDir)),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{currentDir: currentDir, table: t}
	if err := tea.NewProgram(
		&m,
		//tea.WithAltScreen(),
	).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getFileName(fileInfo os.FileInfo) string {
	var dirAffix string
	if fileInfo.IsDir() {
		dirAffix = "/"
	}

	return fileInfo.Name() + dirAffix
}

func getTableRowsFromDir(dir string) []table.Row {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	rows := make([]table.Row, 0, len(entries)+1) // +1 for ..

	rows = append(rows, table.Row{"../", "", ""})

	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			panic(err)
		}

		rows = append(rows, table.Row{getFileName(fileInfo), strconv.FormatInt(fileInfo.Size(), 10), fileInfo.ModTime().String()})
	}

	return rows
}
