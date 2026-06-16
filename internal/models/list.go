package models

import "time"

type Feed struct {
	Title      string
	NyaaURL    string
	TorrentURL string
	Published  time.Time
	Size       string
	Seeders    uint16
	Leechers   uint16
}

type Torrent struct {
	Title      string  `json:"name"`
	Downloaded int64   `json:"downloaded"`
	Size       int64   `json:"size"`
	Progress   float64 `json:"progress"`
	Speed      int64   `json:"dlspeed"`
	Seeders    uint16  `json:"num_seeds"`
	Leechers   uint16  `json:"num_leechs"`
	State      string  `json:"state"`
	Hash       string  `json:"hash"`
}
