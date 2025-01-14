package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	conf "api-pay/config"
	initialization "api-pay/init"
	"api-pay/middleware"
	"api-pay/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	initialization.Initialization()

	port := getPort()

	app := fiber.New(fiber.Config{
		AppName:                 conf.AppConfig.App.Name,
		Prefork:                 conf.AppConfig.App.Prefork,
		BodyLimit:               conf.AppConfig.App.BodyLimit * 1024 * 1024,
		EnableTrustedProxyCheck: true,
	})

	// 启用 CORS 中间件，放在其他中间件之前
	app.Use(cors.New(cors.Config{
		AllowOrigins:     conf.AppConfig.Cors.AllowOrigins, // 允许的前端域名
		AllowMethods:     conf.AppConfig.Cors.AllowMethods, // 允许的方法
		AllowHeaders:     conf.AppConfig.Cors.AllowHeaders, // 允许的请求头
		AllowCredentials: false,                            // 是否允许发送 Cookie 或凭证
	}))

	// 使用请求日志中间件
	app.Use(middleware.RequestLogger(middleware.RequestLoggerConfig{
		Logger:       initialization.GetCurrentLogger(),
		SkipPaths:    conf.AppConfig.Logging.SkipPaths,
		ExcludePaths: conf.AppConfig.Logging.ExcludePaths,
	}))

	// 初始化IP白名单配置
	ipConfig := conf.NewIPWhitelistConfig()

	// 添加IP白名单中间件 (需要在认证中间件之前)
	app.Use(middleware.IPWhitelistMiddleware(ipConfig))

	routes.InitRoutes(app)

	// 捕获所有未匹配的路由
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(http.StatusNotFound).SendString("Hi - This is a bad request. Please stop accessing it !")
	})

	// 创建通道监听信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR2)

	// 创建关闭通道
	done := make(chan bool, 1)

	go func() {
		for {
			sig := <-sigChan
			fmt.Printf("Received signal: %v on port %d\n", sig, port)

			if sig == syscall.SIGUSR2 {
				// 启动新实例
				if err := startNewInstance(port); err != nil {
					fmt.Printf("Error starting new instance on port %d: %v\n", port, err)
					continue
				}
			} else {
				if err := app.ShutdownWithTimeout(3 * time.Second); err != nil {
					fmt.Printf("Error during shutdown on port %d: %v\n", port, err)
				}
				done <- true
				break
			}
		}
	}()

	// 启动服务器
	fmt.Printf("Starting server on port %d...\n", port)
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		fmt.Printf("Error starting server on port %d: %v\n", port, err)
	}

	<-done // 等待关闭信号
}

// 获取端口配置
func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return conf.AppConfig.Port
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Invalid port number %s, using default port %d\n", portStr, conf.AppConfig.Port)
		return conf.AppConfig.Port
	}

	return port
}

func startNewInstance(currentPort int) error {
	fmt.Printf("Starting new instance from port %d...\n", currentPort)
	// 设置环境变量
	env := os.Environ()
	newPort := conf.AppConfig.PortBackup
	if currentPort == conf.AppConfig.PortBackup {
		newPort = conf.AppConfig.Port
	}

	env = append(env, fmt.Sprintf("PORT=%d", newPort))

	// 创建新的命令
	cmd := exec.Command(os.Args[0])
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动新进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new instance: %v", err)
	}

	// 等待新实例准备好
	time.Sleep(3 * time.Second)

	// 健康检查
	url := fmt.Sprintf("http://localhost:%d/api/pay/health", newPort)
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			fmt.Printf("New instance started successfully on port %d\n", newPort)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("health check failed for new instance on port %d", newPort)
}
