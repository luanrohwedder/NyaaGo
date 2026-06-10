package cli

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/feed"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

type (
	item struct {
		title    string
		desc     string
		url      string
		seeders  uint16
		leechers uint16
		size     string
	}

	itemDelegate struct{}

	searchResultMsg struct {
		feeds []feed.FeedResults
		query string
		err   error
	}

	model struct {
		textInput textinput.Model
		list      list.Model
		spinner   spinner.Model
		cfg       *config.Config
		qbClient  *torrent.QbittorrentClient
		loading   bool
		status    string
	}
)

var (
	docStyle = lipgloss.NewStyle().
			Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			MarginBottom(1)
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)
	itemTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)
	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114"))
	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) Seeders() uint16     { return i.seeders }
func (i item) Leechers() uint16    { return i.leechers }
func (i item) Size() string        { return i.size }
func (i item) FilterValue() string { return i.title }

func (d itemDelegate) Height() int                               { return 3 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	prefix := "  "
	title := itemTitleStyle.Render(i.title)
	meta := metaStyle.Render(fmt.Sprintf(
		"Seeders: %d  Leechers: %d  Size: %s",
		i.seeders, i.leechers, i.size,
	))
	link := linkStyle.Render(i.desc)

	if index == m.Index() {
		prefix = selectedItemStyle.Render("> ")
		title = selectedItemStyle.Render(i.title)
	}

	fmt.Fprintf(
		w,
		"%s%s\n  %s\n  %s",
		prefix,
		title,
		meta,
		link,
	)
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func searchCmd(cfg *config.Config, query string) tea.Cmd {
	return func() tea.Msg {
		feeds, err := feed.Search(cfg, query)
		return searchResultMsg{
			feeds: feeds,
			query: query,
			err:   err,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.loading {
				return m, nil
			}

			query := strings.TrimSpace(m.textInput.Value())
			m.loading = true
			m.status = ""
			return m, tea.Batch(searchCmd(m.cfg, query), m.spinner.Tick)
		case "up", "down", "pgup", "pgdown":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case "right":
			selected, ok := m.list.SelectedItem().(item)
			if !ok {
				m.status = "No item selected"
				return m, nil
			}
			if m.qbClient == nil || !m.qbClient.Logged {
				m.status = "qBittorrent is not connected"
				return m, nil
			}

			err := m.qbClient.AddTorrent(selected.url)
			if err != nil {
				m.status = fmt.Sprintf("Failed to Add Torrent: %s", err.Error())
				return m, nil
			}

			m.status = "Torrent Added"
			return m, nil
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		width := max(20, msg.Width-h)
		m.textInput.SetWidth(max(10, width-searchBoxStyle.GetHorizontalFrameSize()))
		m.list.SetSize(width, max(8, msg.Height-v-7))

	case searchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Falha ao buscar: %v", msg.err)
			return m, nil
		}

		items := feedItems(msg.feeds)
		cmd := m.list.SetItems(items)
		m.list.ResetSelected()
		if msg.query == "" {
			m.status = fmt.Sprintf("%d resultados recentes", len(items))
		} else {
			m.status = fmt.Sprintf("%d resultados para %q", len(items), msg.query)
		}
		return m, cmd

	case spinner.TickMsg:
		if m.loading {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			return m, spinnerCmd
		}
	}

	var listCmd, inputCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.textInput, inputCmd = m.textInput.Update(msg)
	return m, tea.Batch(listCmd, inputCmd)
}

func (m model) View() tea.View {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("nyaaGO"),
		subtitleStyle.Render("Search for animes in Nyaa"),
	)
	searchBox := searchBoxStyle.Render(m.textInput.View())

	status := m.status
	if m.loading {
		status = fmt.Sprintf("%s Searching...", m.spinner.View())
	}
	if strings.HasPrefix(status, "Fail") {
		status = errorStyle.Render(status)
	} else {
		status = statusStyle.Render(status)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		searchBox,
		status,
		m.list.View(),
		helpStyle.Render("Enter: search  |  Up/Down: navigate  |  Esc: exit"),
	)
	str := docStyle.Render(content)
	v := tea.NewView(str)
	v.AltScreen = true
	return v
}

func feedItems(feeds []feed.FeedResults) []list.Item {
	items := make([]list.Item, 0, len(feeds))
	for _, it := range feeds {
		newItem := item{
			title:    it.Title,
			desc:     it.NyaaURL,
			url:      it.TorrentURL,
			seeders:  it.Seeders,
			leechers: it.Leechers,
			size:     it.Size,
		}
		items = append(items, newItem)
	}
	return items
}

func newModel(cfg *config.Config, feeds []feed.FeedResults, qbClient *torrent.QbittorrentClient) model {
	ti := textinput.New()
	ti.Prompt = "Search > "
	ti.Placeholder = "Ex.: Hikaru no Go"
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(48)
	inputStyles := ti.Styles()
	inputStyles.Focused.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)
	inputStyles.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	ti.SetStyles(inputStyles)

	results := list.New(feedItems(feeds), itemDelegate{}, 0, 0)
	results.Title = "Results"
	results.SetShowTitle(false)
	results.SetShowFilter(false)
	results.SetFilteringEnabled(false)
	results.SetShowHelp(false)
	results.DisableQuitKeybindings()
	results.SetStatusBarItemName("result", "results")
	results.Styles.StatusBar = results.Styles.StatusBar.
		Foreground(lipgloss.Color("245"))
	results.Styles.PaginationStyle = results.Styles.PaginationStyle.
		Foreground(lipgloss.Color("212"))
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = statusStyle

	m := model{
		textInput: ti,
		list:      results,
		spinner:   sp,
		cfg:       cfg,
		qbClient:  qbClient,
		status:    fmt.Sprintf("%d", len(feeds)),
	}
	return m
}

func NewProgram(cfg *config.Config, feeds []feed.FeedResults, qbClient *torrent.QbittorrentClient) *tea.Program {
	return tea.NewProgram(newModel(cfg, feeds, qbClient))
}
