package main

import (
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filePicker struct {
	filePath string
	table    table.Model
}

const dirAffix = "/"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func initFilePicker(dir string) *filePicker {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Size (B)", Width: 15},
		{Title: "Date modified", Width: 30},
	}

	mTable := table.New(
		table.WithColumns(columns),
		table.WithRows(getTableRowsFromDir(dir)),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	mTable.SetStyles(styles)

	return &filePicker{
		filePath: dir,
		table:    mTable,
	}
}

func getFileName(fileInfo os.FileInfo) string {
	if fileInfo.IsDir() {
		return fileInfo.Name() + dirAffix
	}

	return fileInfo.Name()
}

func getTableRowsFromDir(dir string) []table.Row {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	rows := make([]table.Row, 0, len(entries)+1) // +1 for

	rows = append(rows, table.Row{"../", "", ""})

	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			panic(err)
		}

		rows = append(
			rows,
			table.Row{getFileName(fileInfo), strconv.FormatInt(fileInfo.Size(), 10), fileInfo.ModTime().String()},
		)
	}

	return rows
}

func (f *filePicker) navigateToCurrentSelection(msg tea.Msg, filePath string) tea.Cmd {
	f.filePath = filePath
	f.table.SetRows(getTableRowsFromDir(f.filePath))
	f.table.GotoTop()

	return f.Update(msg)
}

func (f *filePicker) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	f.table, cmd = f.table.Update(msg)

	return cmd
}
