package cli

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

var (
	qbConfigStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("212")).
			Bold(true).
			PaddingBottom(1)

	qbConfigLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true).
				Padding(0, 1)

	qbConfigInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1)

	qbConfigFocusedInputStyle = qbConfigInputStyle.
					BorderForeground(lipgloss.Color("205"))

	qbConfigFieldStyle = lipgloss.NewStyle().
				MarginRight(2)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type settingView struct {
	focusIndex int
	inputs     []textinput.Model
	fieldWidth int
	status     string
	cfg        *config.Config
	qbClient   **torrent.QbittorrentClient
}

func newConfigView(cfg *config.Config, qb **torrent.QbittorrentClient) *settingView {
	sv := settingView{
		inputs:     make([]textinput.Model, 3),
		fieldWidth: 24,
		cfg:        cfg,
		qbClient:   qb,
	}

	var t textinput.Model
	for i := range sv.inputs {
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
		sv.inputs[i] = t
	}
	return &sv
}

func (sv settingView) Init() tea.Cmd {
	return textinput.Blink
}

func (sv *settingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+s":
			sv.status = ""

			if sv.inputs[0].Value() == "" {
				sv.status = "Username is empty"
			} else if sv.inputs[1].Value() == "" {
				sv.status = "Password is empty"
			} else if sv.inputs[2].Value() == "" {
				sv.status = "Port is empty"
			}

			if sv.status != "" {
				break
			}

			qbConfig := config.QbittorrentConfig{
				Username: sv.inputs[0].Value(),
				Password: sv.inputs[1].Value(),
				Port:     sv.inputs[2].Value(),
			}

			qbClient, err := torrent.NewQbittorrentClient(qbConfig)
			if err != nil {
				sv.status = err.Error()
				break
			}

			sv.cfg.Qbittorrent = qbConfig
			if err := sv.cfg.UpdateConfig(); err != nil {
				sv.status = err.Error()
				break
			}

			*sv.qbClient = qbClient
			sv.status = "Config saved!"
			return sv, nil
		case "right", "left":
			s := msg.String()

			if s == "right" {
				sv.focusIndex++
			} else if s == "left" {
				sv.focusIndex--
			}

			if sv.focusIndex > len(sv.inputs) {
				sv.focusIndex = 0
			} else if sv.focusIndex < 0 {
				sv.focusIndex = len(sv.inputs)
			}

			cmds := make([]tea.Cmd, len(sv.inputs))
			for i := range sv.inputs {
				if i == sv.focusIndex {
					cmds[i] = sv.inputs[i].Focus()
					continue
				}
				sv.inputs[i].Blur()
			}
			return sv, tea.Batch(cmds...)
		}
	}

	cmds := make([]tea.Cmd, len(sv.inputs))
	for i := range sv.inputs {
		sv.inputs[i], cmds[i] = sv.inputs[i].Update(msg)
	}
	return sv, tea.Batch(cmds...)
}

func (sv settingView) View() tea.View {
	qbContent := sv.renderQbConfig()

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		sv.status,
		qbContent,
	)

	return tea.NewView(content)
}

func (sv settingView) renderQbConfig() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		qbConfigStyle.Render("🌐 qBittorrent"),
	)
	labels := []string{"Username", "Password", "Port"}
	inputs := make([]string, 0, len(sv.inputs))

	for index, input := range sv.inputs {
		style := qbConfigInputStyle.Width(sv.fieldWidth)
		if index == sv.focusIndex {
			style = qbConfigFocusedInputStyle.Width(sv.fieldWidth)
		}

		field := lipgloss.JoinVertical(
			lipgloss.Left,
			qbConfigLabelStyle.Render(labels[index]),
			style.Render(input.View()),
		)

		fieldStyle := qbConfigFieldStyle
		if index == len(sv.inputs)-1 {
			fieldStyle = fieldStyle.MarginRight(0)
		}
		inputs = append(inputs, fieldStyle.Render(field))
	}

	inputsArea := lipgloss.JoinHorizontal(lipgloss.Top, inputs...)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		inputsArea,
	)
}

func (sv *settingView) setSize(width, height int) {
	const fieldGap = 2

	sv.fieldWidth = max(10, (width-fieldGap*(len(sv.inputs)-1))/len(sv.inputs))
	inputWidth := max(1, sv.fieldWidth-2)

	for i := range sv.inputs {
		sv.inputs[i].SetWidth(inputWidth)
	}
}

func (sv settingView) getHelper() string {
	return "ctrl+s: save"
}
