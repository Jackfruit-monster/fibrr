package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"api-pay/db"
	initialization "api-pay/init"
	"api-pay/utils"
	"github.com/gofiber/fiber/v2"
)

// CreateOrder 回调请求结构
type CreateOrder struct {
	UserId        string  `json:"user_id"`
	Item          string  `json:"item"`
	ItemId        string  `json:"item_id"`
	SinglePric    float64 `json:"single_pric"`
	AmountNum     int64   `json:"amount_num"`
	ServerFlag    string  `json:"server_flag"`
	Description   string  `json:"description"`
	GameRoleId    string  `json:"game_role_id"`
	GameRoleName  string  `json:"game_role_name"`
	GameRoleGrade string  `json:"game_role_grade"`
}

// HandleCreateOrder 处理回调请求
func HandleCreateOrder(c *fiber.Ctx) error {
	resp := utils.NewResponse(c)
	var req CreateOrder

	// 解析请求体
	if err := c.BodyParser(&req); err != nil {
		return resp.Fail(fiber.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err))
	}

	// 验证必要字段
	if req.UserId == "" || req.Item == "" || req.SinglePric <= 0 {
		return resp.Fail(fiber.StatusBadRequest, "Missing required fields")
	}

	// 查询商品是否存在
	gameGoods, err := db.GetOrderByGoodsIitem(req.Item, req.SinglePric)
	if err != nil {
		return resp.Fail(fiber.StatusBadRequest, "商品不存在，或者价格不正确")
	}

	// 先查询是否已存在未支付的订单
	existingOrder, err := db.GetOrderByUserAndItem(req.UserId, req.Item, req.SinglePric)
	if err != nil {
		return resp.Fail(fiber.StatusBadRequest, "查询订单失败")
	}
	if existingOrder != nil {
		// 返回已存在的未支付订单
		return resp.SuccessWithData(&fiber.Map{
			"message":     "订单存在未支付的订单",
			"user_id":     existingOrder.UserId,
			"item":        existingOrder.Item,
			"order":       existingOrder.Order,
			"single_pric": existingOrder.SinglePrice,
		})
	}

	// 创建订单记录
	gameOrder := db.GameOrder{
		UserId:        req.UserId,
		Item:          req.Item,
		ItemId:        gameGoods.ID,
		Order:         initialization.SnowFlake.NextID("811"),
		SinglePrice:   req.SinglePric,
		AmountNum:     req.AmountNum,
		ServerFlag:    req.ServerFlag,
		Description:   req.Description,
		GameRoleId:    req.GameRoleId,
		GameRoleName:  req.GameRoleName,
		GameRoleGrade: req.GameRoleGrade,
		Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
	}

	// 保存到数据库
	if err := db.InsertOrder(&gameOrder); err != nil {
		return resp.Fail(fiber.StatusInternalServerError, "保存失败，已经存在或者数据不正确")
	}

	// 返回成功响应
	return resp.SuccessWithData(&fiber.Map{
		"message":     "订单已创建",
		"user_id":     gameOrder.UserId,
		"item":        gameOrder.Item,
		"item_id":     gameOrder.ItemId,
		"order":       gameOrder.Order,
		"single_pric": gameOrder.SinglePrice,
	})
}

// cancel
type CancelOrder struct {
	UserId      string  `json:"user_id"`
	Item        string  `json:"item"`
	SinglePric  float64 `json:"single_pric"`
	Order       string  `json:"order"`
	Description string  `json:"description"`
}

// HandleCancelOrder
func HandleCancelOrder(c *fiber.Ctx) error {
	resp := utils.NewResponse(c)
	var req CancelOrder

	// 解析请求体
	if err := c.BodyParser(&req); err != nil {
		return resp.Fail(fiber.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err))
	}

	// 验证必要字段
	if req.UserId == "" || req.Order == "" {
		return resp.Fail(fiber.StatusBadRequest, "Missing required fields")
	}

	// 验证订单是否存在
	isExists, err := db.GetOrderByNo(req.Order)
	if err != nil {
		return resp.FailWithCode(fiber.StatusInternalServerError, "查询订单错误", "DATA_ERROR")
	}
	if !isExists {
		return resp.FailWithCode(fiber.StatusInternalServerError, "不存在未支付订单", "DATA_ERROR")
	}

	// 查询这个订单是否支付成功
	isSuccess, _ := db.GetOrderPayExistsByOrderNo(req.Order)
	if isSuccess {
		return resp.FailWithCode(fiber.StatusInternalServerError, "订单已支付，无法取消", "DATA_ERROR")
	}

	// 取消该用户的订单
	err = db.UpdateOrderStatusCancel(req.UserId, req.Order)
	if err != nil {
		return resp.SuccessWithData(&fiber.Map{
			"message": fmt.Sprintf("Failed to cancel order: %v", err),
		})
	}

	// 返回成功响应
	return resp.SuccessWithData(&fiber.Map{
		"message": "订单已取消",
		"user_id": req.UserId,
		"order":   req.Order,
	})
}

// CallbackRequest 回调请求结构
type CallbackRequest struct {
	GameOrderNo   string  `json:"game_order_no"`
	GyyxOrderNo   string  `json:"gyyx_order_no"`
	Result        string  `json:"result"`
	ResultMessage string  `json:"result_message"`
	RmbYuan       float64 `json:"rmb_yuan"`
	ServerFlag    string  `json:"server_flag"`
	CommonParam   string  `json:"common_param"`
	Timestamp     string  `json:"timestamp"`
	Sign          string  `json:"sign"`
	SignType      string  `json:"sign_type"`
}

// HandleCallback 处理回调请求
func HandleCallback(c *fiber.Ctx) error {
	resp := utils.NewResponse(c)

	// 解析 URL 查询参数
	req := CallbackRequest{
		GameOrderNo:   c.Query("game_order_no"),
		GyyxOrderNo:   c.Query("gyyx_order_no"),
		Result:        c.Query("result"),
		ResultMessage: c.Query("result_message"),
		RmbYuan:       c.QueryFloat("rmb_yuan"),
		ServerFlag:    c.Query("server_flag"),
		CommonParam:   c.Query("common_param"),
		Timestamp:     c.Query("timestamp"),
		Sign:          c.Query("sign"),
		SignType:      c.Query("signType"),
	}

	// 验证必填字段
	if req.GameOrderNo == "" || req.GyyxOrderNo == "" {
		return resp.FailWithCode(fiber.StatusBadRequest, "Missing required fields", "INVALID_PARAMS")
	}

	// 查询是否存在待支付的订单
	gameOrder, err := db.GetOrderByNoPrice(req.GameOrderNo, req.RmbYuan)
	if err != nil {
		return resp.FailWithCode(fiber.StatusInternalServerError, "未找到未支付的订单或金额有误", "DATA_ERROR")
	}

	// 查询这个单号、价格的订单是否存在
	exists, _ := db.GetOrderPayExists(req.GameOrderNo, req.RmbYuan)
	if exists {
		return resp.FailWithCode(fiber.StatusBadRequest, "订单号已经支付成功", "ORDER_EXISTS")
	}

	// 构建订单对象
	order := db.GameOrderPay{
		UserId:        gameOrder.UserId,
		ItemId:        gameOrder.ItemId,
		Item:          gameOrder.Item,
		GameOrderNo:   req.GameOrderNo,
		GyyxOrderNo:   req.GyyxOrderNo,
		Result:        req.Result,
		ResultMessage: req.ResultMessage,
		RmbYuan:       req.RmbYuan,
		ServerFlag:    req.ServerFlag,
		CommonParam:   req.CommonParam,
		Timestamp:     req.Timestamp,
	}

	// 保存到数据库
	if err := db.InsertOrder(&order); err != nil {
		return resp.FailWithCode(fiber.StatusInternalServerError, "Database error", "DB_ERROR")
	}

	// 更新订单的状态
	if err := db.UpdateOrderStatus(gameOrder.UserId, req.GameOrderNo); err != nil {
		return resp.FailWithCode(fiber.StatusInternalServerError, "Database error", "DB_ERROR")
	}

	// 返回成功响应
	return resp.Success()
}

// GoodsInfo 商品信息结构
type GoodsInfo struct {
	Id         uint    `json:"id"`
	Item       string  `json:"item"`
	SinglePric float64 `json:"single_pric"`
}

// HandleGoods 处理获取商品信息请求
func HandleGoods(c *fiber.Ctx) error {
	response := utils.NewResponse(c)

	// 获取商品ID参数
	merchandiseIDStr := c.Query("id")
	if merchandiseIDStr == "" {
		return response.Fail(fiber.StatusBadRequest, "商品ID不能为空")
	}

	// 将 merchandiseID 转换为 int 类型
	merchandiseID, err := strconv.Atoi(merchandiseIDStr)
	if err != nil {
		return response.Fail(fiber.StatusBadRequest, "商品ID必须是有效的整数")
	}

	// 获取商品信息
	merchandise, err := getGoodsInfo(merchandiseID)
	if err != nil {
		return response.Fail(fiber.StatusInternalServerError, "获取商品信息失败")
	}

	// 返回商品信息
	return response.SuccessWithData(merchandise)
}

// getGoodsInfo 获取商品信息
func getGoodsInfo(id int) (*GoodsInfo, error) {
	gameGoodsDb, err := db.GetOrderByGoodsId(id)
	if err != nil {
		return nil, fmt.Errorf("无法获取商品ID %d 的信息: %w", id, err) // 添加上下文信息
	}

	// 返回商品信息
	return &GoodsInfo{
		Id:         gameGoodsDb.ID,
		Item:       gameGoodsDb.Item,
		SinglePric: gameGoodsDb.SinglePric,
	}, nil
}

// CallbackRequest 飞猪验证
type VerificationRequest struct {
	UserId string `json:"user_id"`
	Item   string `json:"item"`
	ItemId string `json:"item_id"`
	Order  string `json:"order"`
}

type UserVerificationCount struct {
	ID                uint      `gorm:"primaryKey;comment:主键ID"`               // 主键ID
	UserId            string    `gorm:"size:255;not null;comment:用户ID，唯一标识玩家"` // 用户ID
	Item              string    `gorm:"size:255;not null;comment:商品属性"`        // 商品属性
	VerificationCount int       `gorm:"not null;comment:验证次数"`                 // 验证次数
	LastVerifiedAt    time.Time `gorm:"comment:最后验证时间"`                        // 最后验证时间
	CreatedAt         time.Time `gorm:"autoCreateTime;comment:创建时间"`           // 创建时间
	UpdatedAt         time.Time `gorm:"autoUpdateTime;comment:更新时间"`           // 更新时间
}

// HandleVerification 处理回调请求
func HandleVerification(c *fiber.Ctx) error {
	resp := utils.NewResponse(c)
	var req VerificationRequest

	// 解析请求体
	if err := c.BodyParser(&req); err != nil {
		return resp.Fail(fiber.StatusBadRequest, "Invalid request format")
	}

	// 验证必填字段
	if req.UserId == "" || req.Item == "" {
		return resp.FailWithCode(fiber.StatusBadRequest, "Missing required fields", "INVALID_PARAMS")
	}

	gamrOrderPay, err := db.GetOrderPayByUserId(req.UserId, req.Item)
	if err != nil {
		return resp.FailWithCode(fiber.StatusInternalServerError, "Failed to get order by user id", "DATA_ERROR")
	}

	// 返回成功响应
	return resp.SuccessWithData(&fiber.Map{
		"purchase_time": gamrOrderPay.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// HandleSubmitOrder 提交订单
func HandleSubmitOrder(c *fiber.Ctx) error {
	resp := utils.NewResponse(c)

	// 创建HTTPClient实例
	client := utils.NewHTTPClient("https://api.deepseek.com")

	// 设置API Key
	client.SetHeader("Authorization", "Bearer <DeepSeek API Key>")

	// 创建请求体（使用map[string]interface{}）
	request := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Hello!"},
		},
		"stream": false,
	}

	// 调用API（响应体使用map[string]interface{}）
	var response map[string]interface{}
	err := client.Post("/chat/completions", request, &response)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %v", err)
		return resp.Fail(fiber.StatusInternalServerError, "Failed to call API")
	}

	// 打印响应
	fmt.Printf("Response: %+v\n", response)

	// 返回成功响应
	return resp.SuccessWithData(response)
}
