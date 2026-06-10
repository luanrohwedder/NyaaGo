package feed

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/luanrohwedder/nyaa-GO/internal/config"
	"github.com/mmcdole/gofeed"
)

type FeedResults struct {
	Title      string
	NyaaURL    string
	TorrentURL string
	Published  time.Time
	Size       string
	Seeders    uint16
	Leechers   uint16
}

func Feed(cfg *config.Config) ([]FeedResults, error) {
	return Search(cfg, "")
}

func Search(cfg *config.Config, query string) ([]FeedResults, error) {
	feedURL, err := searchURL(cfg.Feeder.BaseURL, query)
	if err != nil {
		return nil, err
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	res, err := parser(feed)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func searchURL(baseURL, query string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	params := u.Query()
	query = strings.TrimSpace(query)
	if query == "" {
		params.Del("q")
	} else {
		params.Set("q", query)
	}
	u.RawQuery = params.Encode()

	return u.String(), nil
}

func parser(feed *gofeed.Feed) ([]FeedResults, error) {
	res := make([]FeedResults, 0)

	for _, it := range feed.Items {
		seeders, err := strconv.Atoi(it.Extensions["nyaa"]["seeders"][0].Value)
		if err != nil {
			return nil, err
		}

		leechers, err := strconv.Atoi(it.Extensions["nyaa"]["leechers"][0].Value)
		if err != nil {
			return nil, err
		}

		newItem := FeedResults{
			Title:      it.Title,
			NyaaURL:    it.GUID,
			TorrentURL: it.Link,
			Published:  *it.PublishedParsed,
			Size:       it.Extensions["nyaa"]["size"][0].Value,
			Seeders:    uint16(seeders),
			Leechers:   uint16(leechers),
		}
		res = append(res, newItem)
	}

	return res, nil
}
