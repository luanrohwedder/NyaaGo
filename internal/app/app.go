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

	var qbClient *torrent.QbittorrentClient

	p := cli.NewConfigProgram(cfg, &qbClient)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	if qbClient == nil || !qbClient.Logged {
		return
	}

	feeds, err := feed.Feed(cfg)
	if err != nil {
		log.Fatal(err)
	}

	p = cli.NewProgram(cfg, feeds, qbClient)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
