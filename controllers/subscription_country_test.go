package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFilterSubscriptionNodesByCountry_NoParamsKeepsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/static/test", nil)

	nodes := []SubscriptionNode{
		{Remark: "US-1", Country: "us"},
		{Remark: "HK-1", Country: "hk"},
		{Remark: "NoTag", Country: ""},
	}

	got := filterSubscriptionNodesByCountry(c, nodes)
	if len(got) != len(nodes) {
		t.Fatalf("len(filtered) = %d, want %d", len(got), len(nodes))
	}
	for i, n := range got {
		if n.Remark != nodes[i].Remark {
			t.Fatalf("node[%d].Remark = %q, want %q", i, n.Remark, nodes[i].Remark)
		}
	}
}

func TestFilterSubscriptionNodesByCountry_IncludeOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/static/test?include=us,jp", nil)

	nodes := []SubscriptionNode{
		{Remark: "US-1", Country: "us"},
		{Remark: "HK-1", Country: "hk"},
		{Remark: "NoTag", Country: ""},
	}

	got := filterSubscriptionNodesByCountry(c, nodes)
	if len(got) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(got))
	}
	if got[0].Remark != "US-1" {
		t.Fatalf("filtered[0].Remark = %q, want US-1", got[0].Remark)
	}
}

func TestFilterSubscriptionNodesByCountry_ExcludeOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/static/test?exclude=hk", nil)

	nodes := []SubscriptionNode{
		{Remark: "US-1", Country: "us"},
		{Remark: "HK-1", Country: "hk"},
		{Remark: "NoTag", Country: ""},
	}

	got := filterSubscriptionNodesByCountry(c, nodes)
	if len(got) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(got))
	}
	for _, n := range got {
		if n.Country == "hk" {
			t.Fatalf("unexpected HK node in filtered result: %+v", n)
		}
	}
}

func TestFilterSubscriptionNodesByCountry_IncludeAndExclude(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/static/test?include=us,hk&exclude=hk", nil)

	nodes := []SubscriptionNode{
		{Remark: "US-1", Country: "us"},
		{Remark: "HK-1", Country: "hk"},
		{Remark: "NoTag", Country: ""},
	}

	got := filterSubscriptionNodesByCountry(c, nodes)
	if len(got) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(got))
	}
	if got[0].Remark != "US-1" {
		t.Fatalf("filtered[0].Remark = %q, want US-1", got[0].Remark)
	}
}

func TestNormalizeAndValidateCountryCode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantError bool
	}{
		{name: "empty allowed", input: "", wantCode: "", wantError: false},
		{name: "valid lower", input: "us", wantCode: "us", wantError: false},
		{name: "valid upper with space", input: "  HK ", wantCode: "hk", wantError: false},
		{name: "invalid length", input: "usa", wantError: true},
		{name: "unsupported code", input: "zz", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAndValidateCountryCode(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("error presence = %v, want %v (err=%v)", err != nil, tt.wantError, err)
			}
			if !tt.wantError && got != tt.wantCode {
				t.Fatalf("normalizeAndValidateCountryCode(%q) = %q, want %q", tt.input, got, tt.wantCode)
			}
		})
	}
}
