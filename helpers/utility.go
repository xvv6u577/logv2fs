package helper

import (
	"context"
	"net/http"
	"time"
)

func CountNodesInUse(nodeStatus map[string]bool) int {
	trueCount := 0
	for _, status := range nodeStatus {
		if status {
			trueCount++
		}
	}
	return trueCount
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
