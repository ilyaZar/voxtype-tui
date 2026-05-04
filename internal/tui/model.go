package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaZar/voxtype-tui/internal/theme"
	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

type Model struct {
	languages []voxtype.Language
	selected  map[string]bool
	cursor    int
	colors    theme.Colors
	done      bool
	cancelled bool
	message   string
}

func New(selected []string, colors theme.Colors) Model {
	selectedMap := make(map[string]bool, len(selected))
	for _, code := range selected {
		selectedMap[code] = true
	}
	return Model{
		languages: voxtype.Languages,
		selected:  selectedMap,
		colors:    colors,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.languages)-1 {
				m.cursor++
			}
		case " ", "tab", "x":
			m.toggleCurrent()
		case "ctrl+a":
			for _, language := range m.languages {
				m.selected[language.Code] = true
			}
			m.message = ""
		case "enter":
			if len(m.SelectedCodes()) == 0 {
				m.message = "Select at least one language."
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.Accent))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.Foreground))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.SelectionForeground)).Background(lipgloss.Color(m.colors.SelectionBackground))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.Yellow))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var out strings.Builder
	out.WriteString(headerStyle.Render("Voxtype languages:"))
	out.WriteString("\n")
	for index, language := range m.languages {
		cursor := "  "
		if index == m.cursor {
			cursor = "> "
		}
		prefix := "[ ] "
		if m.selected[language.Code] {
			prefix = "[x] "
		}
		line := fmt.Sprintf("%s%s%s", cursor, prefix, language.Menu)
		if m.selected[language.Code] {
			line = selectedStyle.Render(line)
		} else {
			line = itemStyle.Render(line)
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	if m.message != "" {
		out.WriteString(warnStyle.Render(m.message))
		out.WriteString("\n")
	}
	out.WriteString("\n")
	out.WriteString(mutedStyle.Render("x toggle  up/down navigate  enter submit  ctrl+a select all"))
	return out.String()
}

func (m Model) Done() bool {
	return m.done
}

func (m Model) Cancelled() bool {
	return m.cancelled
}

func (m Model) SelectedCodes() []string {
	codes := make([]string, 0, len(m.selected))
	for _, language := range m.languages {
		if m.selected[language.Code] {
			codes = append(codes, language.Code)
		}
	}
	return codes
}

func (m *Model) toggleCurrent() {
	code := m.languages[m.cursor].Code
	m.selected[code] = !m.selected[code]
	m.message = ""
}
