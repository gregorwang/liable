package r2

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"comment-review-platform/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Service struct {
	client *s3.Client
	bucket string
}

type VideoMetadata struct {
	Key      string
	Filename string
	Size     int64
	Modified time.Time
}

// NewR2Service creates a new R2 service instance
func NewR2Service() (*R2Service, error) {
	cfg := config.AppConfig

	if cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2BucketName == "" || cfg.R2Endpoint == "" {
		return nil, fmt.Errorf("R2 configuration is incomplete")
	}

	// Create AWS config for R2
	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion("auto"), // R2 uses "auto" region
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"", // session token not needed for R2
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2Endpoint)
		o.UsePathStyle = true // R2 requires path-style URLs
	})

	return &R2Service{
		client: client,
		bucket: cfg.R2BucketName,
	}, nil
}

// GeneratePresignedURL generates a pre-signed URL for video access
func (r *R2Service) GeneratePresignedURL(videoKey string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(videoKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// ListVideosInPath lists all videos in a specific R2 path
func (r *R2Service) ListVideosInPath(prefix string) ([]VideoMetadata, error) {
	var videos []VideoMetadata

	// Ensure prefix ends with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// Common video file extensions
	videoExtensions := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".mkv":  true,
		".webm": true,
		".flv":  true,
		".wmv":  true,
		".m4v":  true,
		".3gp":  true,
	}

	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Skip directories (objects ending with /)
			if strings.HasSuffix(*obj.Key, "/") {
				continue
			}

			// Check if it's a video file
			ext := strings.ToLower(filepath.Ext(*obj.Key))
			if !videoExtensions[ext] {
				continue
			}

			// Extract filename from key
			filename := filepath.Base(*obj.Key)

			videos = append(videos, VideoMetadata{
				Key:      *obj.Key,
				Filename: filename,
				Size:     *obj.Size,
				Modified: *obj.LastModified,
			})
		}
	}

	log.Printf("Found %d videos in path: %s", len(videos), prefix)
	return videos, nil
}

// GetVideoMetadata gets metadata for a specific video
func (r *R2Service) GetVideoMetadata(videoKey string) (*VideoMetadata, error) {
	result, err := r.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(videoKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	filename := filepath.Base(videoKey)

	return &VideoMetadata{
		Key:      videoKey,
		Filename: filename,
		Size:     *result.ContentLength,
		Modified: *result.LastModified,
	}, nil
}

// CheckConnection tests the R2 connection
func (r *R2Service) CheckConnection() error {
	_, err := r.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to R2 bucket: %w", err)
	}

	log.Printf("✅ R2 connection successful, bucket: %s", r.bucket)
	return nil
}

// GetVideoDuration attempts to get video duration (placeholder - would need ffmpeg integration)
func (r *R2Service) GetVideoDuration(videoKey string) (int, error) {
	// This is a placeholder. In a real implementation, you would:
	// 1. Download a small portion of the video
	// 2. Use ffmpeg or similar to extract duration
	// 3. Return the duration in seconds

	// For now, return a default duration based on file size
	metadata, err := r.GetVideoMetadata(videoKey)
	if err != nil {
		return 0, err
	}

	// Rough estimation: assume 1MB = 10 seconds for TikTok videos
	estimatedDuration := int(metadata.Size / (1024 * 1024) * 10)
	if estimatedDuration < 10 {
		estimatedDuration = 10 // minimum 10 seconds
	}
	if estimatedDuration > 60 {
		estimatedDuration = 60 // maximum 60 seconds for TikTok
	}

	return estimatedDuration, nil
}
