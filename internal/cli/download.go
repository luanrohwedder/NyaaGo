package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/cli/components"
	"github.com/luanrohwedder/nyaa-GO/internal/models"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

type itemDownload struct {
	title      string
	downloaded int64
	seeders    uint16
	leechers   uint16
	size       int64
	speed      int64
	progress   float64
	state      string
	hash       string
}

func (i itemDownload) FilterValue() string { return i.title }

type itemDownloadDelegate struct {
	width int
}

func (d itemDownloadDelegate) Height() int                               { return 3 }
func (d itemDownloadDelegate) Spacing() int                              { return 1 }
func (d itemDownloadDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDownloadDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(itemDownload)
	if !ok {
		return
	}

	bar := progress.New(
		progress.WithDefaultBlend(),
		progress.WithWidth(max(10, d.width-4)),
	)

	prefix := "  "
	title := itemTitleStyle.Render(i.title)
	meta := metaStyle.Render(fmt.Sprintf(
		"%s  %s / %s  %s  Seeders: %d  Leechers: %d",
		formatTorrentState(i.state, i.progress),
		formatBytes(i.downloaded),
		formatBytes(i.size),
		formatSpeed(i.speed),
		i.seeders,
		i.leechers,
	))

	if index == m.Index() {
		prefix = selectedItemStyle.Render("> ")
		title = selectedItemStyle.Render(i.title)
	}

	fmt.Fprintf(
		w,
		"%s%s\n  %s\n  %s",
		prefix,
		title,
		bar.ViewAs(i.progress),
		meta,
	)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type torrentsLoadedMsg struct {
	torrents []models.Torrent
	err      error
}

type downloadView struct {
	list       list.Model
	qbClient   **torrent.QbittorrentClient
	status     string
	hasError   bool
	showDialog bool

	confirmDialog components.ConfirmDialog
	selectedHash  string
	deleteFile    bool
}

func newDownloadView(qb **torrent.QbittorrentClient) *downloadView {
	delegate := itemDownloadDelegate{
		width: 80,
	}

	l := list.New([]list.Item{}, delegate, 80, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return &downloadView{
		list:          l,
		qbClient:      qb,
		confirmDialog: components.NewConfirmDialog("Are you sure?", 25, 8, 1),
	}
}

func (dv *downloadView) setDownloads(torrents []models.Torrent) tea.Cmd {
	items := make([]list.Item, 0, len(torrents))

	for _, t := range torrents {
		items = append(items, itemDownload{
			title:      t.Title,
			downloaded: t.Downloaded,
			seeders:    t.Seeders,
			leechers:   t.Leechers,
			size:       t.Size,
			speed:      t.Speed,
			progress:   t.Progress,
			state:      t.State,
			hash:       t.Hash,
		})
	}

	dv.hasError = false
	return dv.list.SetItems(items)
}

func (dv downloadView) fetchTorrentsCmd() tea.Cmd {
	return func() tea.Msg {
		if dv.qbClient == nil || *dv.qbClient == nil {
			return torrentsLoadedMsg{
				err: fmt.Errorf("qbittorrent client is not connected"),
			}
		}

		torrents, err := (*dv.qbClient).GetTorrent("all")
		return torrentsLoadedMsg{
			torrents: torrents,
			err:      err,
		}
	}
}

func (dv downloadView) Init() tea.Cmd {
	return tea.Batch(dv.fetchTorrentsCmd(), tickCmd())
}

func (dv *downloadView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if dv.showDialog {
		confirmMsg := dv.confirmDialog.Update(msg)

		if confirmMsg != nil {
			msg := confirmMsg.(components.ConfirmMsg)
			dv.showDialog = false
			if !msg.Accepted {
				return dv, nil
			}

			err := (*dv.qbClient).RemoveTorrent(dv.selectedHash, dv.deleteFile)
			if err != nil {
				dv.status = err.Error()
				return dv, nil
			}
			return dv, dv.fetchTorrentsCmd()
		}
		return dv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "down":
			var cmd tea.Cmd
			dv.list, cmd = dv.list.Update(msg)
			return dv, cmd
		case "d", "D":
			selected, ok := dv.list.SelectedItem().(itemDownload)
			if !ok {
				return dv, nil
			}

			dv.showDialog = true
			dv.selectedHash = selected.hash
			dv.deleteFile = msg.String() == "D"

			return dv, nil
		}

	case tickMsg:
		return dv, tea.Batch(dv.fetchTorrentsCmd(), tickCmd())

	case torrentsLoadedMsg:
		if msg.err != nil {
			dv.status = msg.err.Error()
			dv.hasError = true
			return dv, nil
		}
		return dv, dv.setDownloads(msg.torrents)
	}

	var cmd tea.Cmd
	dv.list, cmd = dv.list.Update(msg)
	return dv, cmd
}

func (dv downloadView) View() tea.View {
	if len(dv.list.Items()) == 0 {
		status := dv.status
		if dv.hasError {
			status = errorStyle.Render(status)
		} else {
			status = statusStyle.Render(status)
		}
		content := lipgloss.JoinVertical(
			lipgloss.Top,
			status,
		)
		return tea.NewView(content)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		statusStyle.Render(dv.status),
		dv.list.View(),
	)

	if dv.showDialog {
		content = dv.confirmDialog.Render(content)
	}

	return tea.NewView(content)
}

func (dv *downloadView) setSize(width, height int) {
	dv.list.SetSize(max(10, width-20), max(10, height-12))

	delegate := itemDownloadDelegate{
		width: max(10, width-20),
	}
	dv.list.SetDelegate(delegate)

	dx := width/2 - dv.confirmDialog.Width/2
	dy := height/2 - dv.confirmDialog.Height/2
	dv.confirmDialog.SetPosition(dx, dy)
}

func (dv downloadView) getHelper() string {
	return "↑↓: navigate | d: delete | D: delete w/ file"
}

func formatBytes(value int64) string {
	const unit = int64(1024)
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}

	div, exp := unit, 0
	for n := value / unit; n >= unit && exp < 4; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(value)/float64(div), "KMGTPE"[exp])
}

func formatSpeed(speed int64) string {
	if speed <= 0 {
		return "0 B/s"
	}
	return formatBytes(speed) + "/s"
}

func formatTorrentState(state string, torrentProgress float64) string {
	if torrentProgress >= 1 {
		return "Complete"
	}

	switch strings.ToLower(state) {
	case "downloading", "forceddl", "metadl", "forcedmetadl":
		return "Downloading"
	case "pauseddl", "stoppeddl":
		return "Paused"
	case "queued_dl", "queueddl":
		return "Queued"
	case "checkingdl", "checkingup", "checkingresumedata":
		return "Checking"
	case "error", "missingfiles":
		return "Error"
	case "uploading", "forcedup", "stalledup":
		return "Seeding"
	case "pausedup", "stoppedup":
		return "Complete"
	case "stalleddl":
		return "Stalled"
	default:
		if state == "" {
			return "Unknown"
		}
		return state
	}
}
