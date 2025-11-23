package redis

import (
    "context"
    "crypto/tls"
    "log"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

var Client *redis.Client

// InitRedis initializes Redis connection with optional TLS
func InitRedis(addr, password string, db int, useTLS bool, tlsSkipVerify bool) (*redis.Client, error) {
    options := &redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           db,
        Network:      "tcp",
        DialTimeout:  5 * time.Second,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
    }

	// Enable TLS if required (for Upstash)
    if useTLS {
        host := addr
        if i := strings.LastIndex(addr, ":"); i > 0 {
            host = addr[:i]
        }
        options.TLSConfig = &tls.Config{
            MinVersion: tls.VersionTLS12,
            ServerName: host,
            InsecureSkipVerify: tlsSkipVerify,
        }
    }

    if !useTLS {
        if strings.HasPrefix(addr, "localhost") || strings.HasPrefix(addr, "127.0.0.1") {
            options.Network = "tcp4"
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

