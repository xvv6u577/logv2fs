package helper

import (
	"log"
	"net"
	"os"
	"strings"

	localSanitize "github.com/mrz1836/go-sanitize"
)

func SanitizeStr(str string) string {
	return localSanitize.Custom(str, `[^\p{Han}a-zA-Z0-9-._]+`)
}

func CurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// IsIPv6 检测字符串是否为IPv6地址
func IsIPv6(ip string) bool {
	// 解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查是否为IPv6地址（IPv4地址的To4()方法返回非nil，IPv6地址返回nil）
	return parsedIP.To4() == nil
}

// FormatIPForURL 格式化IP地址用于URL，IPv6地址会被方括号包围
func FormatIPForURL(ip string) string {

	// 检查是否为IPv6地址，如果是则添加方括号
	if IsIPv6(ip) {
		return "[" + ip + "]"
	}

	return ip
}
