package routes

import (
	"api-pay/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func InitRoutes(app *fiber.App) {
	fz_pay := app.Group("/api/pay")
	// 获取商品
	fz_pay.Get("/goods", handlers.HandleGoods)
	// 飞猪回调
	fz_pay.Post("/callback", handlers.HandleCallback)
	// 建单接口
	fz_pay.Post("/create-order", handlers.HandleCreateOrder)
	// 删单接口
	fz_pay.Post("/cancel-order", handlers.HandleCancelOrder)
	// 验证接口
	fz_pay.Post("/verification", handlers.HandleVerification)
	// 提交订单
	fz_pay.Post("/submit-order", handlers.HandleSubmitOrder)
	// 系统接口-接口文档
	fz_pay.Get("/api", handlers.HandleApiDoc)
	// 系统接口-指标接口
	fz_pay.Get("/metrics", monitor.New(monitor.Config{Title: "Service Metrics Page"}))
	// 系统接口-健康检查
	fz_pay.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })
}
