# NyaaGO — Nyaa torrents in your CLI

A terminal user interface (TUI) built with Bubble Tea for searching anime, manga, and other content from the Nyaa torrent feed, then sending downloads directly to qBittorrent.

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat&logo=go)](https://go.dev)

![Example](./demo.gif)

## Requirements

* Go 1.26.4
* qBittorrent with Web UI enabled

## Configuration

Create a `config.yaml` file with your qBittorrent client information.

You can also edit these settings directly inside the app.

## Build and run

```sh
go build -o nyaa-go ./cmd/nyaa-go
./nyaa-go
```

## Disclaimer

This is a Go study project. I do not intend to actively maintain it.
