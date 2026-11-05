package app

import (
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// publicBase is the address links for people outside the household start
// with: CALIO_FEED_URL when the feed and shared documents are published with
// Tailscale Funnel (public is then true), or else the address the app was
// opened on, which works only on the tailnet.
func publicBase(e *core.RequestEvent) (base string, public bool) {
	if base := strings.TrimRight(os.Getenv("CALIO_FEED_URL"), "/"); base != "" {
		return base, true
	}
	if origin := e.Request.Header.Get("Origin"); origin != "" {
		return origin, false
	}
	scheme := "http"
	if e.IsTLS() || e.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + e.Request.Host, false
}
