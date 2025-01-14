// init/logger.go
package initialization

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func InitService() {

}

func InitLogger() {
	// 确保日志目录存在
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("create log directory failed: %v", err))
	}

	// 获取当前日期作为日志文件名
	currentTime := time.Now()
	logFileName := fmt.Sprintf("%s/runtime_%s.log", logDir, currentTime.Format("20060102"))

	// 配置 lumberjack
	writer := &lumberjack.Logger{
		Filename:   logFileName, // 日志文件路径
		MaxSize:    500,         // 每个文件最大尺寸，单位是 MB
		MaxBackups: 3,           // 保留的旧文件个数
		MaxAge:     28,          // 保留的天数
		Compress:   true,        // 是否压缩旧文件
		LocalTime:  true,        // 使用本地时间
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建自定义的 core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		zap.InfoLevel,
	)

	// 添加文件名和行号
	Logger = zap.New(core, zap.AddCaller())
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func GetCurrentLogger() *zap.Logger {
	// 获取当前日期
	currentTime := time.Now()
	currentDate := currentTime.Format("20060102")

	// 检查是否需要创建新的日志文件
	logFileName := fmt.Sprintf("logs/runtime_%s.log", currentDate)
	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		// 重新初始化 logger
		InitLogger()
	}

	return Logger
}
