package cli

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/feed"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

var (
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
)

type itemResult struct {
	title    string
	desc     string
	url      string
	seeders  uint16
	leechers uint16
	size     string
}

func (i itemResult) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 3 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(itemResult)
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

type searchMsg struct {
	feeds []feed.FeedResults
	query string
	err   error
}

type searchView struct {
	textInput textinput.Model
	list      list.Model
	cfg       *config.Config
	qbClient  *torrent.QbittorrentClient
	status    string
	loading   bool
	width     int
	height    int
}

func newSearchView(cfg *config.Config, feeds []feed.FeedResults, qbClient *torrent.QbittorrentClient) *searchView {
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

	results := list.New(convertFeedItems(feeds), itemDelegate{}, 0, 0)
	results.Title = "Results"
	results.SetShowTitle(false)
	results.SetShowFilter(false)
	results.SetFilteringEnabled(false)
	results.SetShowHelp(false)
	results.SetShowStatusBar(false)
	results.DisableQuitKeybindings()
	results.SetStatusBarItemName("result", "results")
	results.Styles.StatusBar = results.Styles.StatusBar.
		Foreground(lipgloss.Color("245"))
	results.Styles.PaginationStyle = results.Styles.PaginationStyle.
		Foreground(lipgloss.Color("212"))

	sv := searchView{
		list:      results,
		cfg:       cfg,
		qbClient:  qbClient,
		status:    "",
		textInput: ti,
	}
	return &sv
}

func convertFeedItems(feeds []feed.FeedResults) []list.Item {
	items := make([]list.Item, 0, len(feeds))
	for _, it := range feeds {
		newItem := itemResult{
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

func (sv *searchView) Init() tea.Cmd {
	return textinput.Blink
}

func searchCmd(cfg *config.Config, query string) tea.Cmd {
	return func() tea.Msg {
		feeds, err := feed.Search(cfg, query)
		return searchMsg{
			feeds: feeds,
			query: query,
			err:   err,
		}
	}
}

func (sv *searchView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "down":
			var cmd tea.Cmd
			sv.list, cmd = sv.list.Update(msg)
			return sv, cmd
		case "enter":
			if sv.loading {
				return sv, nil
			}

			query := strings.TrimSpace(sv.textInput.Value())
			sv.loading = true
			sv.status = ""
			return sv, searchCmd(sv.cfg, query)
		case "right":
			selected, ok := sv.list.SelectedItem().(itemResult)
			if !ok {
				sv.status = "No item selected"
				return sv, nil
			}
			if sv.qbClient == nil || !sv.qbClient.Logged {
				sv.status = "qBittorrent is not connected"
				return sv, nil
			}

			err := sv.qbClient.AddTorrent(selected.url)
			if err != nil {
				sv.status = fmt.Sprintf("Failed to Add Torrent: %s", err.Error())
				return sv, nil
			}

			sv.status = "Torrent Added"
			return sv, nil
		}

	case searchMsg:
		sv.loading = false
		if msg.err != nil {
			sv.status = fmt.Sprintf("Failed to seach: %v", msg.err)
			return sv, nil
		}

		items := convertFeedItems(msg.feeds)
		cmd := sv.list.SetItems(items)
		sv.list.ResetSelected()
		if msg.query == "" {
			sv.status = fmt.Sprintf("%d recents results", len(items))
		} else {
			sv.status = fmt.Sprintf("%d results for %q", len(items), msg.query)
		}
		return sv, cmd
	}

	var listCmd, inputCmd tea.Cmd
	sv.list, inputCmd = sv.list.Update(msg)
	sv.textInput, inputCmd = sv.textInput.Update(msg)
	return sv, tea.Batch(listCmd, inputCmd)
}

func (sv searchView) View() tea.View {
	searchBox := searchBoxStyle.Render(sv.textInput.View())
	status := sv.status
	if strings.HasPrefix(status, "Fail") {
		status = errorStyle.Render(status)
	} else {
		status = statusStyle.Render(status)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		searchBox,
		status,
		sv.list.View(),
	)
	return tea.NewView(content)
}

func (sv *searchView) setSize(width, height int) {
	sv.width = width
	sv.height = height

	sv.textInput.SetWidth(max(10, width/2))
	sv.list.SetSize(max(10, width-20), max(10, height-12))
}

func (sv searchView) getHelper() string {
	return "↑↓: navigate | →: download"
}
