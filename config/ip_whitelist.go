package conf

import (
	"api-pay/utils"
	"github.com/gofiber/fiber/v2"
)

// IPWhitelistConfig IP白名单配置
type IPWhitelistConfig struct {
	AllowedIPs   []string      // 允许访问的IP列表
	IncludePaths []string      // 不需要IP验证的路径
	ExcludePaths []string      // 不需要IP验证的路径
	ErrorHandler fiber.Handler // 自定义错误处理
}

// DefaultIPWhitelistConfig 默认配置
var DefaultIPWhitelistConfig = IPWhitelistConfig{
	AllowedIPs:   AppConfig.IPWhitelist.AllowedIPs,   // 默认允许的IP
	IncludePaths: AppConfig.IPWhitelist.IncludePaths, // 默认不排除的路径
	ExcludePaths: AppConfig.IPWhitelist.ExcludePaths, // 默认排除的路径
	ErrorHandler: func(c *fiber.Ctx) error {
		resp := utils.NewResponse(c)
		return resp.FailWithCode(
			fiber.StatusForbidden,
			"此时暂时不能访问", // state 字段用于显示错误信息
			"403001",   // errorCode 字段
		)
	},
}

// NewIPWhitelistConfig 创建新的IP白名单配置
func NewIPWhitelistConfig() IPWhitelistConfig {
	return GetDefaultIPWhitelistConfig()
}

// GetDefaultIPWhitelistConfig 返回最新的IP白名单配置
func GetDefaultIPWhitelistConfig() IPWhitelistConfig {
	return IPWhitelistConfig{
		AllowedIPs:   AppConfig.IPWhitelist.AllowedIPs,
		IncludePaths: AppConfig.IPWhitelist.IncludePaths,
		ExcludePaths: AppConfig.IPWhitelist.ExcludePaths,
		ErrorHandler: func(c *fiber.Ctx) error {
			resp := utils.NewResponse(c)
			return resp.FailWithCode(
				fiber.StatusForbidden,
				"此时暂时不能访问",
				"403001",
			)
		},
	}
}
