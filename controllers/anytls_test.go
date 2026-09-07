package controllers

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestBuildAnyTLSShareLink(t *testing.T) {
	node := SubscriptionNode{
		Remark:      "Tokyo AnyTLS",
		Domain:      "fallback.example.com",
		IP:          "2001:db8::1",
		SNI:         "edge.example.com",
		SERVER_PORT: "9094",
	}

	link := buildAnyTLSShareLink(node, "password with space")
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse generated AnyTLS link: %v", err)
	}
	if parsed.Scheme != "anytls" {
		t.Fatalf("scheme = %q, want anytls", parsed.Scheme)
	}
	if parsed.Host != "[2001:db8::1]:9094" {
		t.Fatalf("host = %q, want IPv6 host with port", parsed.Host)
	}
	if parsed.User.Username() != "password with space" {
		t.Fatalf("password was not encoded and decoded losslessly")
	}
	if parsed.Fragment != node.Remark {
		t.Fatalf("fragment = %q, want %q", parsed.Fragment, node.Remark)
	}
	const wantRawQuery = "peer=edge.example.com&udp=1&keepalive=30,30,5&fingerprint=Chrome"
	if parsed.RawQuery != wantRawQuery {
		t.Fatalf("raw query = %q, want %q", parsed.RawQuery, wantRawQuery)
	}

	query := parsed.Query()
	wants := map[string]string{
		"peer":        node.SNI,
		"udp":         "1",
		"keepalive":   "30,30,5",
		"fingerprint": "Chrome",
	}
	for key, want := range wants {
		if got := query.Get(key); got != want {
			t.Errorf("query %s = %q, want %q", key, got, want)
		}
	}
}

func TestBuildAnyTLSShareLinkFallsBackToDomainForSNI(t *testing.T) {
	node := SubscriptionNode{
		Remark:      "AnyTLS",
		Domain:      "node.example.com",
		IP:          "192.0.2.10",
		SERVER_PORT: "443",
	}

	parsed, err := url.Parse(buildAnyTLSShareLink(node, "secret"))
	if err != nil {
		t.Fatalf("parse generated AnyTLS link: %v", err)
	}
	if got := parsed.Query().Get("peer"); got != node.Domain {
		t.Fatalf("peer = %q, want domain fallback %q", got, node.Domain)
	}
}

func TestAnyTLSJSONOmitsInsecureInReleaseMode(t *testing.T) {
	tests := []struct {
		name         string
		ginMode      string
		wantInsecure bool
	}{
		{name: "release", ginMode: "release", wantInsecure: false},
		{name: "debug", ginMode: "debug", wantInsecure: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GIN_MODE", test.ginMode)

			outbound := AnyTLSJSON{}
			outbound.TLS.Insecure = getSkipCertVerify()
			encoded, err := json.Marshal(outbound)
			if err != nil {
				t.Fatalf("marshal AnyTLS outbound: %v", err)
			}

			hasInsecure := strings.Contains(string(encoded), `"insecure"`)
			if hasInsecure != test.wantInsecure {
				t.Fatalf("encoded JSON = %s, insecure presence = %t, want %t", encoded, hasInsecure, test.wantInsecure)
			}
		})
	}
}
