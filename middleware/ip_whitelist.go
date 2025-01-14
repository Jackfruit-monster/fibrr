package middleware

import (
	"net"
	"strings"

	"api-pay/config"
	"github.com/gofiber/fiber/v2"
)

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(config conf.IPWhitelistConfig) fiber.Handler {
	// 如果没有配置,使用默认配置
	if config.ErrorHandler == nil {
		config.ErrorHandler = conf.DefaultIPWhitelistConfig.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// 提取 IP 部分，去掉端口号
		clientIP := GetClientIP(c)

		// 检查IP是否在白名单中
		allowed := false
		for _, ip := range config.AllowedIPs {
			if ip == clientIP {
				allowed = true
				break
			}
		}

		// 检查排除路径
		pathExcluded := false
		path := c.Path()
		for _, excludePath := range config.ExcludePaths {
			// 如果路径是通配符格式，检查前缀
			if strings.HasSuffix(excludePath, "/*") {
				basePath := strings.TrimSuffix(excludePath, "/*")
				if strings.HasPrefix(path, basePath) {
					pathExcluded = true
					break
				}
			} else if path == excludePath {
				// 精确匹配
				pathExcluded = true
				break
			}
		}

		// 检查包含路径
		for _, includePath := range config.IncludePaths {
			if path == includePath {
				pathExcluded = false
				break
			}
		}

		if !allowed && pathExcluded {
			return config.ErrorHandler(c)
		}

		return c.Next()
	}
}

// GetClientIP 安全地获取客户端 IP 地址
func GetClientIP(c *fiber.Ctx) string {
	// 按优先级获取 IP
	// 1. 优先使用 X-Real-IP
	if realIP := c.Get("X-Real-IP"); realIP != "" {
		if ip := parseIP(realIP); ip != "" {
			return ip
		}
	}

	// 2. 尝试 X-Forwarded-For
	if forwardedIP := c.Get("X-Forwarded-For"); forwardedIP != "" {
		// 处理可能的多个 IP 情况
		ips := strings.Split(forwardedIP, ",")
		for _, ipStr := range ips {
			if ip := parseIP(strings.TrimSpace(ipStr)); ip != "" {
				return ip
			}
		}
	}

	// 3. 最后使用 RemoteAddr
	remoteAddr := c.Context().RemoteAddr().String()
	if ip := parseIP(remoteAddr); ip != "" {
		return ip
	}

	return ""
}

// parseIP 安全地解析 IP 地址
func parseIP(ipStr string) string {
	// 如果包含端口，先分割
	if strings.Contains(ipStr, ":") {
		ipStr = strings.Split(ipStr, ":")[0]
	}

	// 验证 IP 地址的有效性
	if ip := net.ParseIP(ipStr); ip != nil {
		return ipStr
	}

	return ""
}
