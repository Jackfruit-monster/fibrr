package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient 是一个通用的HTTP客户端
type HTTPClient struct {
	BaseURL    string            // 基础URL
	Headers    map[string]string // 默认请求头
	HTTPClient *http.Client      // HTTP客户端实例
}

// NewHTTPClient 创建一个新的HTTPClient实例
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		Headers:    make(map[string]string),
		HTTPClient: &http.Client{},
	}
}

// SetHeader 设置请求头
func (c *HTTPClient) SetHeader(key, value string) {
	c.Headers[key] = value
}

// Post 发送POST请求
func (c *HTTPClient) Post(endpoint string, body interface{}, response interface{}) error {
	// 构造URL
	url := c.BaseURL + endpoint

	// 将请求体转换为JSON
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Get 发送GET请求
func (c *HTTPClient) Get(endpoint string, response interface{}) error {
	// 构造URL
	url := c.BaseURL + endpoint

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
