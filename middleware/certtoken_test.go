package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCertTokenRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/cert-targets", CertTargetsToken(), func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestCertTargetsToken(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		header map[string]string
		want   int
	}{
		{"env empty rejects", "", map[string]string{CertTargetsTokenHeader: "anything"}, http.StatusServiceUnavailable},
		{"missing token", "s3cret-token", nil, http.StatusUnauthorized},
		{"wrong token", "s3cret-token", map[string]string{CertTargetsTokenHeader: "wrong"}, http.StatusUnauthorized},
		{"prefix of token", "s3cret-token", map[string]string{CertTargetsTokenHeader: "s3cret"}, http.StatusUnauthorized},
		{"header ok", "s3cret-token", map[string]string{CertTargetsTokenHeader: "s3cret-token"}, http.StatusOK},
		{"bearer ok", "s3cret-token", map[string]string{"Authorization": "Bearer s3cret-token"}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CERT_TARGETS_TOKEN", tt.env)
			req := httptest.NewRequest(http.MethodGet, "/v1/cert-targets", nil)
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			newCertTokenRouter().ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}
