package helper

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	// 移除可能的端口号
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		// 检查是否是端口号（不是IPv6地址的一部分）
		portPart := ip[colonIndex+1:]
		if _, err := strconv.Atoi(portPart); err == nil {
			ip = ip[:colonIndex]
		}
	}

	// 解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查是否为IPv6地址
	return parsedIP.To4() == nil
}

// FormatIPForURL 格式化IP地址用于URL，IPv6地址会被方括号包围
func FormatIPForURL(ip string) string {
	// 如果已经是方括号格式，直接返回
	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		return ip
	}

	// 检查是否包含端口号
	colonIndex := strings.LastIndex(ip, ":")
	if colonIndex != -1 {
		// 尝试解析端口号
		portPart := ip[colonIndex+1:]
		if _, err := strconv.Atoi(portPart); err == nil {
			// 有端口号，分离IP和端口
			ipPart := ip[:colonIndex]
			if IsIPv6(ipPart) {
				return "[" + ipPart + "]:" + portPart
			}
			return ip // IPv4地址保持不变
		}
	}

	// 没有端口号的情况
	if IsIPv6(ip) {
		return "[" + ip + "]"
	}

	return ip
}
