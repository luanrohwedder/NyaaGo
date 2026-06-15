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

	feeds, err := feed.Search(cfg, "")
	if err != nil {
		log.Fatal(err)
	}

	var qbClient *torrent.QbittorrentClient

	p := cli.New(cfg, feeds, &qbClient)
	if err := cli.Run(p); err != nil {
		log.Fatal(err)
	}
}
