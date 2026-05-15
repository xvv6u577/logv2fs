package singbox

import "testing"

func TestParseControlListenPort(t *testing.T) {
	tests := []struct {
		name   string
		listen string
		want   string
	}{
		{name: "all ipv4", listen: "0.0.0.0:8379", want: "8379"},
		{name: "loopback ipv4", listen: "127.0.0.1:8479", want: "8479"},
		{name: "ipv6 any", listen: "[::]:8379", want: "8379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseControlListenPort(tt.listen)
			if err != nil {
				t.Fatalf("parseControlListenPort(%q) returned error: %v", tt.listen, err)
			}
			if got != tt.want {
				t.Fatalf("parseControlListenPort(%q) = %q, want %q", tt.listen, got, tt.want)
			}
		})
	}
}
