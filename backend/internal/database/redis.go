package database

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(rawURL string) (*redis.Client, error) {
	// URL-encode the password if it contains special characters
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// Try to fix common issue: unencoded password with special chars
		// Format: redis://:password@host:port
		if idx := strings.Index(rawURL, "://:"); idx >= 0 {
			rest := rawURL[idx+4:]
			if atIdx := strings.LastIndex(rest, "@"); atIdx >= 0 {
				password := rest[:atIdx]
				hostPart := rest[atIdx+1:]
				encoded := url.UserPassword("", password)
				rawURL = rawURL[:idx] + "://" + encoded.String() + "@" + hostPart
			}
		}
	} else if parsedURL.User != nil {
		// Re-encode to ensure special chars in password are handled
		if pw, ok := parsedURL.User.Password(); ok {
			parsedURL.User = url.UserPassword("", pw)
			rawURL = parsedURL.String()
		}
	}

	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
