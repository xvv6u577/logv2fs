package main

import (
	"github.com/caster8013/logv2rayfullstack/model"
)

type User = model.User

func main() {
	var input = User{
		Status: "plain",
		NodeInUseStatus: map[string]bool{
			"www.google.com":   true,
			"www.facebook.com": true,
			"www.youtube.com":  false,
		},
	}

	var submittedNodes = map[string]string{
		"google":    "www.google.com",
		"youtube":   "www.youtube.com",
		"twitter":   "www.twitter.com",
		"instagram": "www.instagram.com",
		"reddit":    "www.reddit.com",
	}

	input.ProduceNodeInUse(submittedNodes)
	// pirnt input.NodeInUseStatus
	for node, status := range input.NodeInUseStatus {
		println(node, status)
	}
}
