package conf

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port       int    `yaml:"port"`
	PortBackup int    `yaml:"port_backup"`
	BotKey     string `yaml:"bot_key"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	Database struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`

	App struct {
		Name      string `yaml:"name"`
		Prefork   bool   `yaml:"prefork"`
		BodyLimit int    `yaml:"body_limit"`
	} `yaml:"app"`

	Logging struct {
		SkipPaths    []string `yaml:"skip_paths"`
		ExcludePaths []string `yaml:"exclude_paths"`
	} `yaml:"logging"`

	Cors struct {
		AllowHeaders string `yaml:"allowed_headers"`
		AllowMethods string `yaml:"allowed_methods"`
		AllowOrigins string `yaml:"allowed_origins"`
	} `yaml:"cors"`

	IPWhitelist struct {
		AllowedIPs   []string `yaml:"allowed_ips"`
		IncludePaths []string `yaml:"include_paths"`
		ExcludePaths []string `yaml:"exclude_paths"`
	} `yaml:"ip_whitelist"`
}

type Schema struct {
	Path string `yaml:"path" json:"path"`
}

type SkuList struct {
	SkuID      string   `yaml:"sku_id" json:"skuId"`
	Price      int      `yaml:"price" json:"price"`
	Quantity   int      `yaml:"quantity" json:"quantity"`
	Title      string   `yaml:"title" json:"title"`
	ImageList  []string `yaml:"image_list" json:"imageList"`
	Type       int      `yaml:"type" json:"type"`
	TagGroupID string   `yaml:"tag_group_id" json:"tagGroupId"`
}

var AppConfig Config

// LoadConfig 从 config.yaml 文件加载配置
func LoadConfig() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}

// GetDBConnectionString 获取数据库连接字符串
func GetDBConnectionString() string {
	return AppConfig.Database.User + ":" +
		AppConfig.Database.Password + "@tcp(" +
		AppConfig.Database.Host + ":" +
		fmt.Sprintf("%d", AppConfig.Database.Port) + ")/" +
		AppConfig.Database.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}
