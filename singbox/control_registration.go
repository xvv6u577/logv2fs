package singbox

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
)

// RegisterControlPortFromEnv records this singbox process control port on its
// active subscription_nodes rows so httpserver can discover per-node ports.
func RegisterControlPortFromEnv() {
	domain := strings.TrimSpace(os.Getenv("CURRENT_DOMAIN"))
	if domain == "" {
		log.Printf("[singbox.control_registration] skip: CURRENT_DOMAIN is empty")
		return
	}

	listen := strings.TrimSpace(os.Getenv("SINGBOX_CONTROL_LISTEN"))
	port, err := parseControlListenPort(listen)
	if err != nil {
		log.Printf("[singbox.control_registration] skip: invalid SINGBOX_CONTROL_LISTEN=%q: %v", listen, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"domain": domain,
		"status": "active",
		"type":   bson.M{"$in": []string{"reality", "hysteria2", "vmessws"}},
	}
	update := bson.M{"$set": bson.M{"control_port": port}}

	res, err := database.GetCollection(model.SubscriptionNode{}).UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("[singbox.control_registration] update subscription_nodes failed domain=%s port=%s err=%v",
			domain, port, err)
		return
	}

	log.Printf("[singbox.control_registration] registered control_port=%s domain=%s matched=%d modified=%d",
		port, domain, res.MatchedCount, res.ModifiedCount)
}

func parseControlListenPort(listen string) (string, error) {
	listen = strings.TrimSpace(listen)
	if listen == "" {
		return "", fmt.Errorf("empty listen address")
	}
	_, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "", err
	}
	if port == "" {
		return "", fmt.Errorf("empty port")
	}
	return port, nil
}
