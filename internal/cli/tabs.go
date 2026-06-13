package cli

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/feed"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)
)

type tabModel struct {
	tabs      []string
	activeTab int
	width     int
	height    int

	configModel configModel
	listModel   listModel
}

func (m tabModel) Init() tea.Cmd {
	return nil
}

func (m tabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil
		case "shift+tab":
			m.activeTab = max(0, m.activeTab-1)
			return m, nil
		}
	}

	var cmd tea.Cmd

	switch m.activeTab {
	case 0:
		var mdl tea.Model
		mdl, cmd = m.listModel.Update(msg)
		m.listModel = mdl.(listModel)
	case 2:
		var mdl tea.Model
		mdl, cmd = m.configModel.Update(msg)
		m.configModel = mdl.(configModel)
	}

	return m, cmd
}

func (m *tabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.listModel.SetSize(width, height)
}

func (m tabModel) View() tea.View {
	highlight := lipgloss.Color("#7D56F4")
	out := []string{}

	for i, t := range m.tabs {
		if i == m.activeTab {
			out = append(out, activeTab.BorderForeground(highlight).Render(t))
		} else {
			out = append(out, tab.BorderForeground(highlight).Render(t))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)
	gap := tabGap.BorderForeground(highlight).Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	var content, footer string

	switch m.activeTab {
	case 0:
		content = m.listModel.View().Content
		footer = helpStyle.Width(m.width).Render("enter: search | 🢁🡻: navigate | ➜: download | esc: exit")
	case 1:
		content = "Downloads"
	case 2:
		content = m.configModel.View().Content
		footer = helpStyle.Width(m.width).Render("enter: confirm | 🢁🡻: navigate | esc: exit")
	}

	content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - lipgloss.Height(header) - lipgloss.Height(footer)).
		Render(content)
	content = lipgloss.JoinVertical(lipgloss.Top, header, content, footer)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func NewTabProgram(cfg *config.Config, feeds []feed.FeedResults, qb **torrent.QbittorrentClient) *tea.Program {
	list := newModel(cfg, feeds, *qb)
	config := initConfigModel(cfg, qb)

	m := tabModel{tabs: []string{"List", "Downloads", "Config"}, listModel: list, configModel: config}
	return tea.NewProgram(m)
}
