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

type OutboundsSelector struct {
	Tag       string   `json:"tag"`
	Type      string   `json:"type"`
	Default   string   `json:"default"`
	Outbounds []string `json:"outbounds"`
}

type SingboxJSON struct {
	DNS struct {
		Servers []struct {
			Tag     string `json:"tag"`
			Address string `json:"address"`
			Detour  string `json:"detour,omitempty"`
		} `json:"servers"`
		Rules []struct {
			Outbound     []string `json:"outbound,omitempty"`
			Server       string   `json:"server"`
			DisableCache bool     `json:"disable_cache,omitempty"`
			Geosite      string   `json:"geosite,omitempty"`
			ClashMode    string   `json:"clash_mode,omitempty"`
		} `json:"rules"`
		Strategy string `json:"strategy"`
	} `json:"dns"`
	Inbounds     []interface{} `json:"inbounds"`
	Experimental struct {
		ClashAPI struct {
			ExternalController string `json:"external_controller"`
			Secret             string `json:"secret"`
			StoreSelected      bool   `json:"store_selected"`
		} `json:"clash_api"`
	} `json:"experimental"`
	Log struct {
		Disabled  bool   `json:"disabled"`
		Level     string `json:"level"`
		Timestamp bool   `json:"timestamp"`
	} `json:"log"`
	Outbounds []interface{} `json:"outbounds"`
	Route     struct {
		AutoDetectInterface bool `json:"auto_detect_interface"`
		Rules               []struct {
			Geosite   string   `json:"geosite,omitempty"`
			Outbound  string   `json:"outbound"`
			Protocol  string   `json:"protocol,omitempty"`
			ClashMode string   `json:"clash_mode,omitempty"`
			Geoip     []string `json:"geoip,omitempty"`
		} `json:"rules"`
		Geoip struct {
			DownloadDetour string `json:"download_detour"`
		} `json:"geoip"`
		Geosite struct {
			DownloadDetour string `json:"download_detour"`
		} `json:"geosite"`
	} `json:"route"`
}

type SingboxYAML struct {
	Port                    int    `yaml:"port"`
	AllowLan                bool   `yaml:"allow-lan"`
	Mode                    string `yaml:"mode"`
	LogLevel                string `yaml:"log-level"`
	UnifiedDelay            bool   `yaml:"unified-delay"`
	GlobalClientFingerprint string `yaml:"global-client-fingerprint"`
	Ipv6                    bool   `yaml:"ipv6"`
	DNS                     struct {
		Enable            bool     `yaml:"enable"`
		Listen            string   `yaml:"listen"`
		Ipv6              bool     `yaml:"ipv6"`
		EnhancedMode      string   `yaml:"enhanced-mode"`
		FakeIPRange       string   `yaml:"fake-ip-range"`
		DefaultNameserver []string `yaml:"default-nameserver"`
		Nameserver        []string `yaml:"nameserver"`
		Fallback          []string `yaml:"fallback"`
		FallbackFilter    struct {
			Geoip     bool     `yaml:"geoip"`
			GeoipCode string   `yaml:"geoip-code"`
			Ipcidr    []string `yaml:"ipcidr"`
		} `yaml:"fallback-filter"`
	} `yaml:"dns"`
	Proxies []struct {
		Name              string `yaml:"name"`
		Type              string `yaml:"type"`
		Server            string `yaml:"server"`
		Port              int    `yaml:"port"`
		UUID              string `yaml:"uuid,omitempty"`
		Network           string `yaml:"network,omitempty"`
		UDP               bool   `yaml:"udp,omitempty"`
		TLS               bool   `yaml:"tls,omitempty"`
		Flow              string `yaml:"flow,omitempty"`
		Servername        string `yaml:"servername,omitempty"`
		ClientFingerprint string `yaml:"client-fingerprint,omitempty"`
		RealityOpts       struct {
			PublicKey string `yaml:"public-key"`
			ShortID   string `yaml:"short-id"`
		} `yaml:"reality-opts,omitempty"`
		Password       string   `yaml:"password,omitempty"`
		Sni            string   `yaml:"sni,omitempty"`
		SkipCertVerify bool     `yaml:"skip-cert-verify,omitempty"`
		Alpn           []string `yaml:"alpn,omitempty"`
	} `yaml:"proxies"`
	ProxyGroups []struct {
		Name      string   `yaml:"name"`
		Type      string   `yaml:"type"`
		Proxies   []string `yaml:"proxies"`
		URL       string   `yaml:"url,omitempty"`
		Interval  int      `yaml:"interval,omitempty"`
		Tolerance int      `yaml:"tolerance,omitempty"`
	} `yaml:"proxy-groups"`
	Rules []string `yaml:"rules"`
}
