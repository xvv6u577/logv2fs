package model

type RealityJSON struct {
	Tag            string `json:"tag"`
	Type           string `json:"type"`
	UUID           string `json:"uuid"`
	ServerPort     int    `json:"server_port"`
	Flow           string `json:"flow"`
	PacketEncoding string `json:"packet_encoding"`
	Server         string `json:"server"`
	TLS            struct {
		Enabled    bool   `json:"enabled"`
		ServerName string `json:"server_name"`
		Utls       struct {
			Enabled     bool   `json:"enabled"`
			Fingerprint string `json:"fingerprint"`
		} `json:"utls"`
		Reality struct {
			Enabled   bool   `json:"enabled"`
			PublicKey string `json:"public_key"`
			ShortID   string `json:"short_id"`
		} `json:"reality"`
	} `json:"tls"`
}

type Hysteria2JSON struct {
	Tag        string `json:"tag"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	UpMbps     int    `json:"up_mbps"`
	DownMbps   int    `json:"down_mbps"`
	Password   string `json:"password"`
	TLS        struct {
		Enabled    bool     `json:"enabled"`
		ServerName string   `json:"server_name"`
		Insecure   bool     `json:"insecure"`
		Alpn       []string `json:"alpn"`
	} `json:"tls"`
}

type CFVlessJSON struct {
	Tag        string `json:"tag"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	UUID       string `json:"uuid"`
	Flow       string `json:"flow"`
	TLS        struct {
		Enabled    bool   `json:"enabled"`
		ServerName string `json:"server_name"`
		Insecure   bool   `json:"insecure"`
		Utls       struct {
			Enabled     bool   `json:"enabled"`
			Fingerprint string `json:"fingerprint"`
		} `json:"utls"`
	} `json:"tls"`
	Multiplex struct {
		Enabled    bool   `json:"enabled"`
		Protocol   string `json:"protocol"`
		MaxStreams int    `json:"max_streams"`
	} `json:"multiplex"`
	PacketEncoding string `json:"packet_encoding"`
	Transport      struct {
		Type    string `json:"type"`
		Path    string `json:"path"`
		Headers struct {
			Host string `json:"Host"`
		} `json:"headers"`
	} `json:"transport"`
}

type SingboxJSON struct {
	Log struct {
		Disabled  bool   `json:"disabled"`
		Level     string `json:"level"`
		Timestamp bool   `json:"timestamp"`
	} `json:"log"`
	DNS          interface{}   `json:"dns"`
	Inbounds     []interface{} `json:"inbounds"`
	Experimental interface{}   `json:"experimental"`
	Outbounds    []interface{} `json:"outbounds"`
	Route        interface{}   `json:"route"`
}

type RealityYAML struct {
	Name              string `yaml:"name"`
	Type              string `yaml:"type"`
	Server            string `yaml:"server"`
	Port              int    `yaml:"port"`
	UUID              string `yaml:"uuid"`
	Network           string `yaml:"network"`
	UDP               bool   `yaml:"udp"`
	TLS               bool   `yaml:"tls"`
	Flow              string `yaml:"flow"`
	Servername        string `yaml:"servername"`
	ClientFingerprint string `yaml:"client-fingerprint"`
	RealityOpts       struct {
		PublicKey string `yaml:"public-key"`
		ShortID   string `yaml:"short-id"`
	} `yaml:"reality-opts"`
}

type Hysteria2YAML struct {
	Name           string   `yaml:"name"`
	Type           string   `yaml:"type"`
	Server         string   `yaml:"server"`
	Port           int      `yaml:"port"`
	Password       string   `yaml:"password"`
	Sni            string   `yaml:"sni"`
	SkipCertVerify bool     `yaml:"skip-cert-verify"`
	Alpn           []string `yaml:"alpn"`
}

type CFVlessYAML struct {
	Type              string `yaml:"type"`
	Name              string `yaml:"name"`
	Server            string `yaml:"server"`
	Port              int    `yaml:"port"`
	UUID              string `yaml:"uuid"`
	Network           string `yaml:"network"`
	TLS               bool   `yaml:"tls"`
	UDP               bool   `yaml:"udp"`
	Servername        string `yaml:"servername"`
	ClientFingerprint string `yaml:"client-fingerprint"`
	WsOpts            struct {
		Path    string `yaml:"path"`
		Headers struct {
			Host string `yaml:"Host"`
		} `yaml:"headers"`
	} `yaml:"ws-opts"`
}

type SingboxYAML struct {
	Port                    int           `yaml:"port"`
	AllowLan                bool          `yaml:"allow-lan"`
	Mode                    string        `yaml:"mode"`
	LogLevel                string        `yaml:"log-level"`
	UnifiedDelay            bool          `yaml:"unified-delay"`
	GlobalClientFingerprint string        `yaml:"global-client-fingerprint"`
	Ipv6                    bool          `yaml:"ipv6"`
	DNS                     interface{}   `yaml:"dns"`
	Proxies                 []interface{} `yaml:"proxies"`
	ProxyGroups             []struct {
		Name      string   `yaml:"name"`
		Type      string   `yaml:"type"`
		Proxies   []string `yaml:"proxies"`
		URL       string   `yaml:"url,omitempty"`
		Interval  int      `yaml:"interval,omitempty"`
		Tolerance int      `yaml:"tolerance,omitempty"`
	} `yaml:"proxy-groups"`
	Rules []string `yaml:"rules"`
}

type ClashYAML struct {
	Port               int           `default:"7890" yaml:"mixed-port"`
	AllowLan           bool          `yaml:"allow-lan"`
	BindAddress        string        `yaml:"bind-address"`
	Mode               string        `yaml:"mode"`
	LogLevel           string        `yaml:"log-level"`
	ExternalController string        `yaml:"external-controller"`
	Dns                interface{}   `yaml:"dns"`
	Proxies            []interface{} `yaml:"proxies"`
	ProxyGroups        []ProxyGroups `yaml:"proxy-groups"`
	Rules              []string      `yaml:"rules"`
}

type Vmess struct {
	Name           string `yaml:"name"`
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Type           string `yaml:"type"`
	UUID           string `yaml:"uuid"`
	AlterID        int    `yaml:"alterId"`
	Cipher         string `yaml:"cipher"`
	TLS            bool   `yaml:"tls"`
	SkipCertVerify bool   `yaml:"skip-cert-verify"`
	Sni            string `yaml:"sni"`
	UDP            bool   `yaml:"udp"`
	Network        string `yaml:"network"`
	WsOpts         struct {
		Path    string `yaml:"path"`
		Headers struct {
			Host string `yaml:"Host"`
		} `yaml:"headers"`
	} `yaml:"ws-opts"`
}
