package cli

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

var (
	configTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)
	configSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	focusedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	configHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type (
	configModel struct {
		focusIndex int
		inputs     []textinput.Model
		cursorMode cursor.Mode
		status     string
		cfg        *config.Config
		qbClient   **torrent.QbittorrentClient
	}
)

func NewConfigProgram(cfg *config.Config, qbClient **torrent.QbittorrentClient) *tea.Program {
	return tea.NewProgram(initConfigModel(cfg, qbClient))
}

func initConfigModel(cfg *config.Config, qb **torrent.QbittorrentClient) configModel {
	m := configModel{
		inputs:   make([]textinput.Model, 3),
		cfg:      cfg,
		qbClient: qb,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 32

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		t.SetStyles(s)

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.SetValue(cfg.Qbittorrent.Username)
			t.Focus()
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.SetValue(cfg.Qbittorrent.Password)
		case 2:
			t.Placeholder = "Port"
			t.CharLimit = 10
			t.SetValue(cfg.Qbittorrent.Port)
		}
		m.inputs[i] = t
	}
	return m
}

func (m configModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				s := m.inputs[i].Styles()
				s.Cursor.Blink = m.cursorMode == cursor.CursorBlink
				m.inputs[i].SetStyles(s)
			}
			return m, tea.Batch(cmds...)
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				if m.inputs[0].Value() == "" {
					m.status = "Username is empty"
				} else if m.inputs[1].Value() == "" {
					m.status = "Password is empty"
				} else if m.inputs[2].Value() == "" {
					m.status = "Port is empty"
				}

				if m.status != "" {
					break
				}

				qbConfig := config.QbittorrentConfig{
					Username: m.inputs[0].Value(),
					Password: m.inputs[1].Value(),
					Port:     m.inputs[2].Value(),
				}

				qbClient, err := torrent.NewQbittorrentClient(qbConfig)
				if err != nil {
					m.status = err.Error()
					break
				}

				m.cfg.Qbittorrent = qbConfig
				if err := m.cfg.UpdateConfig(); err != nil {
					m.status = err.Error()
					break
				}

				*m.qbClient = qbClient
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}
			return m, tea.Batch(cmds...)
		}

	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *configModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m configModel) View() tea.View {
	var b strings.Builder
	var c *tea.Cursor

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		configTitleStyle.Render("NyaaGO"),
		configSubtitleStyle.Render("configure your qbittorrent client"),
	)

	status := m.status

	for i, in := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
		if m.cursorMode != cursor.CursorHide && in.Focused() {
			c = in.Cursor()
			if c != nil {
				c.Y += i
			}
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		status,
		b.String(),
		configHelpStyle.Render("Up/Down: navigate | Esc: exit"),
	)

	v := tea.NewView(content)
	v.Cursor = c
	v.AltScreen = true
	return v
}
