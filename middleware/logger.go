package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"api-pay/init"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestLog 请求日志结构
type RequestLog struct {
	TraceID   string            `json:"trace_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	IP        string            `json:"ip"`
	Status    int               `json:"status"`
	Duration  time.Duration     `json:"duration"`
	StartTime time.Time         `json:"start_time"`
	Response  string            `json:"response"`
}

// RequestLoggerConfig 中间件配置
type RequestLoggerConfig struct {
	Logger       *zap.Logger
	SkipPaths    []string
	ExcludePaths []string
}

// cleanJSON 清理和压缩JSON字符串，去除换行符、多余空格和不必要的转义
func cleanJSON(input string) string {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "{") && !strings.HasPrefix(input, "[") {
		return input
	}

	// 尝试解析JSON到通用接口
	var temp interface{}
	if err := json.Unmarshal([]byte(input), &temp); err != nil {
		return input
	}

	// 重新编码为紧凑的JSON
	clean, err := json.Marshal(temp)
	if err != nil {
		return input
	}

	return string(clean)
}

// RequestLogger 创建请求日志中间件
func RequestLogger(config RequestLoggerConfig) fiber.Handler {
	// 预处理路径映射
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		if strings.HasSuffix(path, "/*") {
			skipPaths[strings.TrimSuffix(path, "/*")] = true
		} else {
			skipPaths[path] = true
		}
	}

	// 预处理强制记录日志的路径
	excludePaths := make(map[string]bool)
	for _, path := range config.ExcludePaths {
		excludePaths[path] = true
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()

		// 判断是否需要跳过日志记录
		if shouldSkipLogging(path, skipPaths) && !excludePaths[path] {
			return c.Next()
		}

		// 获取当前的 logger
		logger := initialization.GetCurrentLogger()

		// 生成 trace_id
		traceID := uuid.New().String()
		c.Locals("trace_id", traceID)
		c.Set("X-Trace-ID", traceID)

		// 记录请求开始时间
		startTime := time.Now()

		// 获取请求头
		headers := extractHeaders(c)

		// 清理请求体的 JSON
		requestBody := cleanJSON(string(c.Body()))

		// 构建请求日志对象
		reqLog := buildRequestLog(traceID, c, headers, requestBody, startTime)

		// 记录请求日志
		logRequest(logger, reqLog)

		// 处理请求
		err := c.Next()

		// 更新并记录响应日志
		updateAndLogResponse(logger, c, &reqLog, startTime)

		return err
	}
}

// 判断是否跳过日志记录的辅助函数
func shouldSkipLogging(path string, skipPaths map[string]bool) bool {
	for basePath := range skipPaths {
		if strings.HasPrefix(path, basePath) {
			return true
		}
	}
	return false
}

// 提取请求头的辅助函数
func extractHeaders(c *fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// 构建请求日志对象的辅助函数
func buildRequestLog(traceID string, c *fiber.Ctx, headers map[string]string,
	requestBody string, startTime time.Time) RequestLog {
	return RequestLog{
		TraceID:   traceID,
		Method:    c.Method(),
		Path:      c.Path(),
		Query:     string(c.Request().URI().QueryString()),
		Headers:   headers,
		Body:      requestBody,
		IP:        c.Get("X-Forwarded-For"),
		StartTime: startTime,
	}
}

// 记录请求日志的辅助函数
func logRequest(logger *zap.Logger, reqLog RequestLog) {
	logger.Info("incoming request",
		zap.String("trace_id", reqLog.TraceID),
		zap.String("method", reqLog.Method),
		zap.String("path", reqLog.Path),
		zap.String("query", reqLog.Query),
		zap.Any("body", reqLog.Body),
		zap.String("ip", reqLog.IP),
		zap.Any("headers", reqLog.Headers),
	)
}

// 更新并记录响应日志的辅助函数
func updateAndLogResponse(logger *zap.Logger, c *fiber.Ctx,
	reqLog *RequestLog, startTime time.Time) {
	responseBody := c.Response().Body()
	if responseBody != nil {
		reqLog.Response = cleanJSON(string(responseBody))
	}

	reqLog.Status = c.Response().StatusCode()
	reqLog.Duration = time.Since(startTime)

	logger.Info("request completed",
		zap.String("trace_id", reqLog.TraceID),
		zap.String("method", reqLog.Method),
		zap.String("path", reqLog.Path),
		zap.Int("status", reqLog.Status),
		zap.Duration("duration", reqLog.Duration),
		zap.Any("response", reqLog.Response),
	)
}
