package redis

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// InitRedis initializes Redis connection with optional TLS
func InitRedis(addr, password string, db int, useTLS bool) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	// Enable TLS if required (for Upstash)
	if useTLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(options)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	Client = client
	log.Println("✅ Redis connected successfully")
	return client, nil
}

// Close closes the Redis connection
func Close() {
	if Client != nil {
		Client.Close()
	}
}

