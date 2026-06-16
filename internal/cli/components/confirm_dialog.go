package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

var (
	selectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(charmtone.Charple).
			Underline(true).
			Align(lipgloss.Center).
			Padding(0, 2)
	unselectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(charmtone.Charcoal).
			Align(lipgloss.Center).
			Padding(0, 2)
)

type ConfirmMsg struct {
	Accepted bool
}

type ConfirmDialog struct {
	Width       int
	Height      int
	Layer       *lipgloss.Layer
	OptionLayer *lipgloss.Layer
	focusIndex  int
}

func NewConfirmDialog(text string, width, height, z int) ConfirmDialog {
	layer := lipgloss.NewLayer(newCard(true, text, width, height)).
		AddLayers().
		X(0).
		Y(0).
		Z(z)
	return ConfirmDialog{
		Width:       width,
		Height:      height,
		Layer:       layer,
		OptionLayer: lipgloss.NewLayer("").Z(z),
	}
}

func (m *ConfirmDialog) Update(msg tea.Msg) tea.Msg {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "right", "left":
			k := msg.String()
			if k == "left" {
				m.focusIndex--
			} else if k == "right" {
				m.focusIndex++
			}

			if m.focusIndex < 0 {
				m.focusIndex = 1
			} else if m.focusIndex > 1 {
				m.focusIndex = 0
			}
		case "enter":
			if m.focusIndex == 0 {
				return ConfirmMsg{Accepted: true}
			} else {
				return ConfirmMsg{Accepted: false}
			}
		}
	}
	return nil
}

func (m ConfirmDialog) Render(base string) string {
	baseLayer := lipgloss.NewLayer(base).X(0).Y(0).Z(0)

	var options string
	if m.focusIndex == 0 {
		options = lipgloss.JoinHorizontal(
			lipgloss.Center,
			selectStyle.Render("Yes"),
			"  ",
			unselectStyle.Render("No"),
		)
	} else {
		options = lipgloss.JoinHorizontal(
			lipgloss.Center,
			unselectStyle.Render("Yes"),
			"  ",
			selectStyle.Render("No"),
		)
	}

	m.OptionLayer = lipgloss.NewLayer(options).
		X(m.Layer.GetX() + max(0, (m.Width-lipgloss.Width(options))/2)).
		Y(m.Layer.GetY() + max(0, m.Height-3)).
		Z(m.OptionLayer.GetZ())

	content := lipgloss.NewCompositor(
		baseLayer,
		m.Layer,
		m.OptionLayer,
	)
	return content.Render()
}

func (m *ConfirmDialog) SetPosition(x, y int) {
	m.Layer.X(x)
	m.Layer.Y(y)
	m.OptionLayer.X(x)
	m.OptionLayer.Y(y)
}

func newCard(darkMode bool, text string, width, height int) string {
	lightDark := lipgloss.LightDark(darkMode)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(
			charmtone.Cherry,
			charmtone.Charple,
			charmtone.Guac,
			charmtone.Charple,
			charmtone.Sriracha,
		).
		Foreground(lightDark(charmtone.Iron, charmtone.Butter)).
		Height(height).
		Width(width).
		Padding(1, 1).
		Align(lipgloss.Center).
		Render(text)
}
