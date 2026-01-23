package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

type ReviewOutput struct {
	IsApproved bool     `json:"is_approved"`
	Tags       []string `json:"tags"`
	Reason     string   `json:"reason"`
	Confidence int      `json:"confidence"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) ReviewComment(ctx context.Context, commentText string, allowedTags []string) (ReviewOutput, string, error) {
	missing := make([]string, 0, 3)
	if c.baseURL == "" {
		missing = append(missing, "base_url")
	}
	if c.apiKey == "" {
		missing = append(missing, "api_key")
	}
	if c.model == "" {
		missing = append(missing, "model")
	}
	if len(missing) > 0 {
		return ReviewOutput{}, "", fmt.Errorf("ai client not configured: missing %s", strings.Join(missing, ", "))
	}

	tagHint := ""
	if len(allowedTags) > 0 {
		tagHint = fmt.Sprintf("可选标签列表：%s。若 is_approved 为 false，优先从列表中选择 1-3 个标签；若没有合适标签，可生成简短中文标签作为补充。若 is_approved 为 true，tags 为空数组。", strings.Join(allowedTags, "、"))
	} else {
		tagHint = "若 is_approved 为 true，tags 为空数组；若 is_approved 为 false，生成 1-3 个简短中文标签。"
	}

	systemPrompt := "你是评论合规审核助手，需要判断评论是否合规。" +
		"仅返回 JSON 对象，包含字段：is_approved（布尔值）、tags（字符串数组）、reason（字符串，中文原因）、confidence（0-100 整数）。" +
		tagHint +
		"置信度说明：50 表示非常不确定，60-70 表示偏不确定，80 表示较确定，90 以上代表高度确定；避免默认输出 85/95。" +
		"不要输出额外文本。"

	userPrompt := fmt.Sprintf("Comment:\n%s", commentText)

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReviewOutput{}, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ReviewOutput{}, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ReviewOutput{}, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		bodyText := strings.TrimSpace(string(bodyBytes))
		if bodyText == "" {
			return ReviewOutput{}, "", fmt.Errorf("ai request failed with status %d", resp.StatusCode)
		}
		return ReviewOutput{}, "", fmt.Errorf("ai request failed with status %d: %s", resp.StatusCode, bodyText)
	}

	var parsedResponse chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResponse); err != nil {
		return ReviewOutput{}, "", fmt.Errorf("ai response decode failed: %w", err)
	}
	if len(parsedResponse.Choices) == 0 {
		return ReviewOutput{}, "", errors.New("ai response missing choices")
	}

	rawContent := strings.TrimSpace(parsedResponse.Choices[0].Message.Content)
	result, err := parseReviewOutput(rawContent)
	if err != nil {
		return ReviewOutput{}, rawContent, fmt.Errorf("ai response parse failed: %w", err)
	}

	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 100 {
		result.Confidence = 100
	}

	return result, rawContent, nil
}

func parseReviewOutput(raw string) (ReviewOutput, error) {
	var output ReviewOutput
	if err := json.Unmarshal([]byte(raw), &output); err == nil {
		return output, nil
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return ReviewOutput{}, errors.New("ai response is not valid json")
	}

	trimmed := raw[start : end+1]
	if err := json.Unmarshal([]byte(trimmed), &output); err != nil {
		return ReviewOutput{}, err
	}

	return output, nil
}
