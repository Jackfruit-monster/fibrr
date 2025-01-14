package initialization

import (
	"log"

	config "api-pay/config"
	"api-pay/db"
	"api-pay/utils"
)

var SnowFlake *utils.Snowflake
var RandString *utils.StringGenerator

func Initialization() {

	// 初始化配置文件
	config.LoadConfig()

	// 初始化日志
	InitLogger()

	RandString = utils.NewStringGenerator()

	// 初始化雪花算法
	SnowFlake = utils.GetSnowflake(1, 1)

	// 初始化数据库
	if err := db.InitDB(); err != nil {
		log.Fatal("Failed to connect to database ")
	}

	//// 初始化Redis
	//if err := db.InitRedis(); err != nil {
	//	log.Fatal("Failed to connect to redis ")
	//}

	//// 初始化机器人
	//wxbot.InitBot()

}
