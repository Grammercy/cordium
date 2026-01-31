package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AllowedDomains restricts which domains can be proxied to prevent misuse
var AllowedDomains = []string{
	"cdn.discordapp.com",
	"media.discordapp.net",
	"discordapp.com",
	"discord.com",
}

func isAllowed(targetURL string) bool {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	hostname := u.Hostname()
	for _, domain := range AllowedDomains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}

func Handler(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// Validate URL
	if !isAllowed(rawURL) {
		http.Error(w, "Domain not allowed", http.StatusForbidden)
		return
	}

	// Fetch remote resource
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(rawURL)
	if err != nil {
		http.Error(w, "Failed to fetch resource", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// Stream body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
