package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
	height    int
}

func New(selected []string, colors theme.Colors, languages ...[]voxtype.Language) Model {
	selectedMap := make(map[string]bool, len(selected))
	for _, code := range selected {
		selectedMap[code] = true
	}
	return Model{
		languages: languageList(languages...),
		selected:  selectedMap,
		colors:    colors,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
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
	background := lipgloss.Color(m.colors.Background)
	baseStyle := lipgloss.NewStyle().Background(background)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.colors.Yellow)).
		Background(background).
		Bold(true).
		Underline(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.Foreground)).Background(background)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.SelectionForeground)).Background(lipgloss.Color(m.colors.SelectionBackground))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.colors.Red)).Background(background)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Background(background)
	footer := mutedStyle.Render("x toggle  up/down navigate  enter submit  ctrl+a select all")
	contentLines := 1 + len(m.languages)
	if m.message != "" {
		contentLines++
	}

	menus := languageMenus(m.languages)
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
		line := fmt.Sprintf("%s%s%s", cursor, prefix, menus[index])
		if m.selected[language.Code] {
			line = selectedStyle.Render(line)
		} else {
			line = itemStyle.Render(line)
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	if m.message != "" {
		out.WriteString(errorStyle.Render(m.message))
		out.WriteString("\n")
	}
	for range footerBlankLines(m.height, contentLines) {
		out.WriteString("\n")
	}
	out.WriteString(footer)
	return baseStyle.Render(out.String())
}

func languageMenus(languages []voxtype.Language) []string {
	width := 0
	for _, language := range languages {
		width = max(width, utf8.RuneCountInString(languageName(language)))
	}

	menus := make([]string, 0, len(languages))
	for _, language := range languages {
		name := languageName(language)
		padding := strings.Repeat(" ", width-utf8.RuneCountInString(name))
		menus = append(menus, fmt.Sprintf("%s%s (%s)", name, padding, languageShortcut(language)))
	}
	return menus
}

func languageName(language voxtype.Language) string {
	if language.Name != "" {
		return language.Name
	}
	return strings.ToUpper(language.Code)
}

func languageShortcut(language voxtype.Language) string {
	shortcut := strings.ToUpper(language.Code)
	if len([]rune(shortcut)) != 2 {
		shortcut = strings.ToUpper(language.Label)
	}
	runes := []rune(shortcut)
	if len(runes) > 2 {
		return string(runes[:2])
	}
	for len(runes) < 2 {
		runes = append(runes, ' ')
	}
	return string(runes)
}

func languageList(languages ...[]voxtype.Language) []voxtype.Language {
	if len(languages) > 0 && len(languages[0]) > 0 {
		return languages[0]
	}
	return nil
}

func footerBlankLines(height int, contentLines int) int {
	if height <= 0 {
		return 1
	}
	blankLines := height - contentLines - 1
	if blankLines < 1 {
		return 1
	}
	return blankLines
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
	if len(m.languages) == 0 {
		return
	}
	code := m.languages[m.cursor].Code
	m.selected[code] = !m.selected[code]
	m.message = ""
}
