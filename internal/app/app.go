package app

import (
	"log"

	"github.com/luanrohwedder/nyaa-GO/internal/cli"
	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/luanrohwedder/nyaa-GO/internal/feed"
	"github.com/luanrohwedder/nyaa-GO/internal/torrent"
)

func Run() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	feeds, err := feed.Feed(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var qbClient *torrent.QbittorrentClient

	p := cli.NewTabProgram(cfg, feeds, &qbClient)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
