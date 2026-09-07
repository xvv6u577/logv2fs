package helper

import "testing"

func TestTLSServerName(t *testing.T) {
	if got := TLSServerName("  sni.example.com  ", "domain.example.com"); got != "sni.example.com" {
		t.Fatalf("explicit SNI = %q", got)
	}
	if got := TLSServerName(" ", "domain.example.com"); got != "domain.example.com" {
		t.Fatalf("fallback SNI = %q", got)
	}
}
