package cli

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/models"
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
)

type Layout struct {
	tabs       []string
	activeView int
	width      int
	height     int
	views      []View
}

func newLayout(cfg *config.Config, feeds []models.Feed, qb **torrent.QbittorrentClient) *Layout {
	layout := Layout{
		tabs:  []string{"Search", "Download", "Settings"},
		views: make([]View, 0),
	}

	layout.views = append(layout.views, newSearchView(cfg, feeds, qb))
	layout.views = append(layout.views, newTorrentView(qb))
	layout.views = append(layout.views, newConfigView(cfg, qb))

	return &layout
}

func (l Layout) Init() tea.Cmd {
	if len(l.views) == 0 {
		return nil
	}
	return l.views[l.activeView].Init()
}

func (l *Layout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return l, tea.Quit
		case "tab":
			l.activeView = (l.activeView + 1) % len(l.views)
			return l, l.views[l.activeView].Init()
		case "shift+tab":
			l.activeView = (l.activeView - 1 + len(l.views)) % len(l.views)
			return l, l.views[l.activeView].Init()
		}
	}

	model, cmd := l.views[l.activeView].Update(msg)

	view, ok := model.(View)
	if ok {
		l.views[l.activeView] = view
	}

	return l, cmd
}

func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height

	for _, t := range l.views {
		t.setSize(width, height)
	}
}

func (l Layout) View() tea.View {
	header := l.renderHeader()
	footer := l.renderFooter()

	content := l.views[l.activeView].View().Content
	content = lipgloss.NewStyle().
		Width(l.width).
		Height(l.height - lipgloss.Height(header) - lipgloss.Height(footer)).
		Render(content)

	layout := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	v := tea.NewView(layout)
	v.AltScreen = true
	return v
}

func (l Layout) renderHeader() string {
	highlight := lipgloss.Color("#7D56F4")
	out := []string{}

	for i, t := range l.tabs {
		if i == l.activeView {
			out = append(out, activeTab.BorderForeground(highlight).Render(t))
		} else {
			out = append(out, tab.BorderForeground(highlight).Render(t))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)
	gap := tabGap.BorderForeground(highlight).Render(strings.Repeat(" ", max(0, l.width-lipgloss.Width(row)-2)))
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	return header
}

func (l Layout) renderFooter() string {
	footer := "⇥: next tab | ⇤: prev tab | " + l.views[l.activeView].getHelper() + " | ESC: exit"
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("144")).
		Render(footer)
}
