package feed

import (
	"net/url"
	"testing"
)

func TestSearchURLSetsQueryParameter(t *testing.T) {
	got, err := searchURL("https://nyaa.si/?page=rss", "Hikaru no Go")
	if err != nil {
		t.Fatalf("searchURL returned an error: %v", err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("could not parse result: %v", err)
	}
	if got := u.Query().Get("page"); got != "rss" {
		t.Fatalf("page = %q, want rss", got)
	}
	if got := u.Query().Get("q"); got != "Hikaru no Go" {
		t.Fatalf("q = %q, want Hikaru no Go", got)
	}
}

func TestSearchURLRemovesEmptyQueryParameter(t *testing.T) {
	got, err := searchURL("https://nyaa.si/?page=rss&q=old", "  ")
	if err != nil {
		t.Fatalf("searchURL returned an error: %v", err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("could not parse result: %v", err)
	}
	if _, ok := u.Query()["q"]; ok {
		t.Fatalf("q should not be present in %q", got)
	}
}
