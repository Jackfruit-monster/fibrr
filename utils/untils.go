package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	// 时间起点 2024-01-01 00:00:00 +0800 CST
	epoch int64 = 1704038400000

	// 机器ID位数
	workerIDBits = 5
	// 数据中心ID位数
	datacenterIDBits = 5
	// 序列号位数
	sequenceBits = 12

	// 最大值
	maxWorkerID     = -1 ^ (-1 << workerIDBits)
	maxDatacenterID = -1 ^ (-1 << datacenterIDBits)
	maxSequence     = -1 ^ (-1 << sequenceBits)

	// 左移位数
	workerIDShift      = sequenceBits
	datacenterIDShift  = sequenceBits + workerIDBits
	timestampLeftShift = sequenceBits + workerIDBits + datacenterIDBits
)

// Snowflake 结构体
type Snowflake struct {
	mutex        sync.Mutex
	timestamp    int64
	workerID     int64
	datacenterID int64
	sequence     int64
}

// 全局单例
var (
	snowflake *Snowflake
	once      sync.Once
)

// GetSnowflake 获取雪花算法单例
func GetSnowflake(datacenterID, workerID int64) *Snowflake {
	once.Do(func() {
		snowflake = NewSnowflake(datacenterID, workerID)
	})
	return snowflake
}

// NewSnowflake 创建一个新的雪花算法实例
func NewSnowflake(datacenterID, workerID int64) *Snowflake {
	// 校验数据中心ID和机器ID
	if datacenterID < 0 || datacenterID > maxDatacenterID {
		panic(fmt.Sprintf("datacenter ID must be between 0 and %d", maxDatacenterID))
	}
	if workerID < 0 || workerID > maxWorkerID {
		panic(fmt.Sprintf("worker ID must be between 0 and %d", maxWorkerID))
	}

	return &Snowflake{
		timestamp:    0,
		datacenterID: datacenterID,
		workerID:     workerID,
		sequence:     0,
	}
}

// NextID 生成下一个ID
func (s *Snowflake) NextID(header string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取当前时间戳
	now := time.Now().UnixMilli()

	// 如果当前时间小于上一次ID生成的时间戳，说明系统时钟回退过
	if now < s.timestamp {
		panic("Clock moved backwards! Refusing to generate ID")
	}

	// 如果是同一时间生成的，则进行序列号自增
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 同一毫秒的序列数已经达到最大
		if s.sequence == 0 {
			// 阻塞到下一个毫秒，获得新的时间戳
			now = s.waitNextMillis(now)
		}
	} else {
		// 不是同一时间，序列号重置为0
		s.sequence = 0
	}

	// 更新上次生成ID的时间戳
	s.timestamp = now

	// 组合ID
	id := ((now - epoch) << timestampLeftShift) |
		(s.datacenterID << datacenterIDShift) |
		(s.workerID << workerIDShift) |
		s.sequence

	return fmt.Sprintf("%s-%d", header, id)
}

// waitNextMillis 等待下一个毫秒
func (s *Snowflake) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixMilli()
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixMilli()
	}
	return timestamp
}

// StringGenerator 用于生成和管理唯一随机字符串
type StringGenerator struct {
	generated map[string]bool
	mutex     sync.RWMutex
}

// NewStringGenerator 创建一个新的字符串生成器实例
func NewStringGenerator() *StringGenerator {
	return &StringGenerator{
		generated: make(map[string]bool),
	}
}

// GenerateUniqueString 生成指定长度的唯一随机字符串
func (g *StringGenerator) GenerateUniqueString(length int, header string) (string, error) {
	if length < 1 {
		return "", errors.New("length must be positive")
	}

	// 设置最大重试次数，防止死循环
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// 生成随机字节
		bytes := make([]byte, length)
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}

		// 使用base64编码，并加入时间戳确保更高的唯一性
		timestamp := time.Now().UnixNano()
		rawString := base64.RawURLEncoding.EncodeToString(bytes)

		// 只取rawString的前length个字符，并将时间戳转为字符串并添加
		uniqueString := rawString[:length] + fmt.Sprintf("%d", timestamp)

		// 检查是否已存在
		g.mutex.Lock()
		if !g.generated[uniqueString] {
			g.generated[uniqueString] = true
			g.mutex.Unlock()
			return fmt.Sprintf("%s-%s", header, uniqueString), nil
		}
		g.mutex.Unlock()
	}

	return "", errors.New("failed to generate unique string after maximum attempts")
}

// IsStringGenerated 检查字符串是否已经生成过
func (g *StringGenerator) IsStringGenerated(s string) bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.generated[s]
}

// Reset 重置生成器状态
func (g *StringGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.generated = make(map[string]bool)
}

func SendRequestJson(url string, requestBody map[string]interface{}, header map[string]string) ([]byte, error) {
	// 将请求数据编码成 JSON
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 遍历并设置额外的请求头
	for key, value := range header {
		req.Header.Set(key, value)
	}

	// 当前时间格式化为 2024-12-22 12:32:12 格式
	formattedTime := time.Now().Format("2006-01-02 15:04:05")

	fmt.Println(fmt.Sprintf("时间：%s HTTP地址:%s 发出的信息：%s", formattedTime, url, string(reqBody)))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)

	fmt.Println(fmt.Sprintf("时间：%s HTTP地址:%s HTTP: 返回的信息：%s", formattedTime, url, string(respBody)))

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %d", resp.StatusCode)
	}

	return respBody, nil
}

// 生成安全的JWT Secret
func GenerateJWTSecret() []byte {
	// 生成32字节(256位)的随机密钥
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		log.Fatal("Failed to generate JWT secret")
	}
	return secret
}
