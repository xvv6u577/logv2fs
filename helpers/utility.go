package helper

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

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

func IsDomainReachable(domain string) bool {
	// Create a context with a timeout of 3 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create an HTTP client with the context deadline.
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+domain, nil)
	if err != nil {
		// fmt.Println("Error:", err)
		return false
	}

	// Perform the HTTP request with the context.
	resp, err := client.Do(req.WithContext(ctx))

	// Check for errors during the request.
	if err != nil {
		// fmt.Println("Error:", err)
		return false
	}

	// Make sure to close the response body to avoid resource leaks.
	defer resp.Body.Close()

	// Check the status code of the response.
	// A status code of 200-299 indicates success (reachable).
	if (resp.StatusCode >= 200 && resp.StatusCode <= 299) || (resp.StatusCode == 400) {
		return true
	}

	return false
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
