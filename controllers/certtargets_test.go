package controllers

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/xvv6u577/logv2fs/model"
)

func TestBuildCertTargets(t *testing.T) {
	nodes := []model.SubscriptionNode{
		{Remark: "LA v4", Domain: "la.example.com", IP: "192.0.2.1", SERVER_PORT: "443", PASSWORD: "p", UUID: "u"},
		{Remark: "LA v6", Domain: "la.example.com", IP: "2001:db8::1", SERVER_PORT: "443"}, // 同 SNI+端口，去重
		{Remark: "LA 8443", Domain: "LA.example.com", IP: "192.0.2.1", SERVER_PORT: "8443"},
		{Remark: "Tokyo", Domain: "fallback.example.com", SNI: " edge.example.com ", IP: "198.51.100.2", SERVER_PORT: ""},
		{Remark: "empty", Domain: "", SNI: ""},
	}

	got := buildCertTargets(nodes)
	want := []CertTarget{
		{SNI: "edge.example.com", IP: "198.51.100.2", Port: "443", Remark: "Tokyo"},
		{SNI: "la.example.com", IP: "192.0.2.1", Port: "443", Remark: "LA v4"},
		{SNI: "la.example.com", IP: "192.0.2.1", Port: "8443", Remark: "LA 8443"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCertTargets() =\n%+v\nwant\n%+v", got, want)
	}

	b, _ := json.Marshal(got)
	for _, forbidden := range []string{"password", "uuid", "public_key", "short_id"} {
		if strings.Contains(strings.ToLower(string(b)), forbidden) {
			t.Fatalf("JSON output leaks %q: %s", forbidden, b)
		}
	}
}
