package model

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestProduceNodeInUse(t *testing.T) {
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
		"facebook":  "www.facebook.com",
		"youtube":   "www.youtube.com",
		"twitter":   "www.twitter.com",
		"instagram": "www.instagram.com",
		"reddit":    "www.reddit.com",
	}

	input.ProduceNodeInUse(submittedNodes)

	assert.Equal(t, len(input.NodeInUseStatus), 6)
	assert.Equal(t, input.NodeInUseStatus["www.google.com"], true)
	assert.Equal(t, input.NodeInUseStatus["www.youtube.com"], false)
	assert.Equal(t, input.NodeInUseStatus["www.reddit.com"], true)
}
