package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"comment-review-platform/internal/models"

	"github.com/redis/go-redis/v9"
)

const metricsEndpointsKey = "metrics:api:endpoints"

type MetricsService struct {
	rdb    *redis.Client
	window time.Duration
}

func NewMetricsService(rdb *redis.Client, windowMinutes int) *MetricsService {
	if windowMinutes <= 0 {
		windowMinutes = 5
	}
	return &MetricsService{
		rdb:    rdb,
		window: time.Duration(windowMinutes) * time.Minute,
	}
}

func (s *MetricsService) Record(method, path string, status int, latency time.Duration) {
	if s == nil || s.rdb == nil {
		return
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	endpoint := fmt.Sprintf("%s %s", method, path)
	_ = s.rdb.SAdd(ctx, metricsEndpointsKey, endpoint).Err()

	bucket := time.Now().UTC().Format("200601021504")
	baseKey := metricsBaseKey(method, path)
	countKey := fmt.Sprintf("%s:counts:%s", baseKey, bucket)
	latencyKey := fmt.Sprintf("%s:latency:%s", baseKey, bucket)
	ttl := s.window + time.Minute

	pipe := s.rdb.Pipeline()
	pipe.HIncrBy(ctx, countKey, "total", 1)
	if status >= 200 && status < 400 {
		pipe.HIncrBy(ctx, countKey, "success", 1)
	} else if status >= 400 && status < 500 {
		pipe.HIncrBy(ctx, countKey, "client_error", 1)
	} else if status >= 500 {
		pipe.HIncrBy(ctx, countKey, "server_error", 1)
	}
	pipe.Expire(ctx, countKey, ttl)
	pipe.RPush(ctx, latencyKey, int(latency.Milliseconds()))
	pipe.Expire(ctx, latencyKey, ttl)
	_, _ = pipe.Exec(ctx)
}

func (s *MetricsService) GetMetrics(windowMinutes int) ([]models.EndpointMetrics, error) {
	if s == nil || s.rdb == nil {
		return nil, nil
	}
	if windowMinutes <= 0 {
		windowMinutes = int(s.window.Minutes())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endpoints, err := s.rdb.SMembers(ctx, metricsEndpointsKey).Result()
	if err != nil {
		return nil, err
	}

	buckets := recentBuckets(windowMinutes)
	metrics := make([]models.EndpointMetrics, 0, len(endpoints))
	for _, endpoint := range endpoints {
		method, path := splitEndpoint(endpoint)
		if method == "" || path == "" {
			continue
		}
		baseKey := metricsBaseKey(method, path)

		var total, success, clientErr, serverErr int
		latencies := make([]int, 0)

		for _, bucket := range buckets {
			countKey := fmt.Sprintf("%s:counts:%s", baseKey, bucket)
			values, _ := s.rdb.HGetAll(ctx, countKey).Result()
			total += parseInt(values["total"])
			success += parseInt(values["success"])
			clientErr += parseInt(values["client_error"])
			serverErr += parseInt(values["server_error"])

			latencyKey := fmt.Sprintf("%s:latency:%s", baseKey, bucket)
			items, _ := s.rdb.LRange(ctx, latencyKey, 0, -1).Result()
			for _, item := range items {
				latencies = append(latencies, parseInt(item))
			}
		}

		avg := averageLatency(latencies)
		p99 := percentileLatency(latencies, 99)
		successRate := 0.0
		if total > 0 {
			successRate = float64(success) / float64(total)
		}

		metrics = append(metrics, models.EndpointMetrics{
			Method:       method,
			Path:         path,
			Total:        total,
			Success:      success,
			ClientError:  clientErr,
			ServerError:  serverErr,
			SuccessRate:  successRate,
			AvgLatencyMs: avg,
			P99LatencyMs: p99,
		})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Total > metrics[j].Total
	})

	return metrics, nil
}

func metricsBaseKey(method, path string) string {
	path = strings.ReplaceAll(path, " ", "_")
	return fmt.Sprintf("metrics:api:%s:%s", method, path)
}

func recentBuckets(windowMinutes int) []string {
	now := time.Now().UTC()
	buckets := make([]string, 0, windowMinutes)
	for i := 0; i < windowMinutes; i++ {
		buckets = append(buckets, now.Add(-time.Duration(i)*time.Minute).Format("200601021504"))
	}
	return buckets
}

func splitEndpoint(endpoint string) (string, string) {
	parts := strings.SplitN(endpoint, " ", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func averageLatency(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0
	for _, v := range values {
		total += v
	}
	return float64(total) / float64(len(values))
}

func percentileLatency(values []int, percentile float64) int {
	if len(values) == 0 {
		return 0
	}
	sort.Ints(values)
	rank := int(math.Ceil(percentile/100*float64(len(values)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(values) {
		rank = len(values) - 1
	}
	return values[rank]
}
