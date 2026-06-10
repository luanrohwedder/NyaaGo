package torrent

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/luanrohwedder/nyaa-GO/internal/config"
)

type QbittorrentClient struct {
	Logged  bool
	client  *http.Client
	baseURL string
}

func NewQbittorrentClient(cfg config.QbittorrentConfig) (*QbittorrentClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	qb := &QbittorrentClient{
		client:  client,
		baseURL: fmt.Sprintf("http://localhost:%s", cfg.Port),
	}

	if err := qb.login(cfg.Username, cfg.Password); err != nil {
		return nil, err
	}
	return qb, nil
}

func (qb *QbittorrentClient) login(username string, password string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(
		http.MethodPost,
		qb.baseURL+"/api/v2/auth/login",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", qb.baseURL)
	req.Header.Set("Origin", qb.baseURL)

	res, err := qb.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("qbittorrent login failed: status=%d body=%q", res.StatusCode, string(body))
	}
	if res.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) != "Ok." {
		return fmt.Errorf("qbittorrent login failed: body=%q", string(body))
	}
	qb.Logged = true
	return nil
}

func (qb *QbittorrentClient) AddTorrent(torrentURL string) error {
	if qb == nil || !qb.Logged || qb.client == nil {
		return fmt.Errorf("qbittorrent client is not connected")
	}

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	if err := w.WriteField("urls", torrentURL); err != nil {
		return err
	}

	if err := w.WriteField("savepath", "/home/luan/Downloads"); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		qb.baseURL+"/api/v2/torrents/add",
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", w.FormDataContentType())
	req.Header.Set("Referer", qb.baseURL)
	req.Header.Set("Origin", qb.baseURL)

	res, err := qb.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("failed to add torrent, status=%d body=%s", res.StatusCode, string(body))
	}
	if strings.EqualFold(strings.TrimSpace(string(body)), "Fails.") {
		return fmt.Errorf("qbittorrent failed to add torrent: body=%q", string(body))
	}
	return nil
}

func (c QbittorrentClient) RemoveTorrent() {

}
