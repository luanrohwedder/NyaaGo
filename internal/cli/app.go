package cli

import (
	tea "charm.land/bubbletea/v2"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/feed"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

type View interface {
	tea.Model
	getHelper() string
	setSize(int, int)
}

func New(cfg *config.Config, feeds []feed.FeedResults, qb **torrent.QbittorrentClient) tea.Model {
	return newLayout(cfg, feeds, qb)
}

func Run(layout tea.Model) error {
	p := tea.NewProgram(layout)
	_, err := p.Run()
	return err
}
