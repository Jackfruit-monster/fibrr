package db

import (
	"time"

	config "api-pay/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type GameGoods struct {
	ID         uint      `gorm:"primaryKey;comment:主键ID"`                  // 主键ID
	Item       string    `gorm:"size:255;uniqueIndex;comment:商品项，表示购买的物品"` // 商品项
	SinglePric float64   `gorm:"type:decimal(10,2);comment:商品价格，保留两位小数"`   // 商品价格
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`              // 创建时间
}

// GameOrder 游戏订单表模型
type GameOrder struct {
	ID            uint       `gorm:"primaryKey;comment:主键ID"` // 主键ID
	UserId        string     `gorm:"size:255;comment:用户ID，唯一标识玩家"`
	Item          string     `gorm:"size:255;comment:商品项，表示购买的物品"`
	ItemId        uint       `gorm:"type:int;comment:商品属性ID"`                 // 商品属性
	SinglePrice   float64    `gorm:"type:decimal(10,2);comment:商品价格，保留两位小数"`  // 商品价格
	OrderStatus   float64    `gorm:"size:255;comment:订单状态 0 未支付 1 已取消 2 已支付"` // 商品价格
	AmountNum     int64      `gorm:"type:int;comment:购买数量"`                   // 购买数量
	Order         string     `gorm:"size:255;comment:游戏订单号，用于标识该订单"`          // 游戏订单号
	ServerFlag    string     `gorm:"size:100;comment:服务器标识，区分订单所属服务器"`        // 服务器标识
	Description   string     `gorm:"size:255;comment:订单描述，描述订单详细信息"`          // 订单描述
	GameRoleId    string     `gorm:"size:255;comment:游戏角色ID"`                 // 游戏角色ID
	GameRoleName  string     `gorm:"size:255;comment:游戏角色名称"`                 // 游戏角色名称
	GameRoleGrade string     `gorm:"size:255;comment:游戏角色等级"`                 // 游戏角色等级
	GameOrderNo   string     `gorm:"size:255;comment:游戏订单号"`                  // 游戏订单号
	GyyxOrderNo   string     `gorm:"size:255;comment:平台订单号"`                  // 平台订单号
	Timestamp     string     `gorm:"size:50;comment:时间戳"`                     // 时间戳
	DeletedAt     *time.Time `gorm:"comment:删除时间，记录删除时间戳"`                    // 删除时间
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间"`             // 创建时间
}

// GameOrderPay 游戏支付成功数据
type GameOrderPay struct {
	ID            uint      `gorm:"primaryKey;comment:主键ID"`                       // 主键ID
	UserId        string    `gorm:"size:255;not null;comment:用户ID，唯一标识玩家"`         // 用户ID
	ItemId        uint      `gorm:"type:int;comment:商品属性ID"`                       // 商品属性
	Item          string    `gorm:"size:255;not null;comment:商品属性"`                // 商品属性
	GameOrderNo   string    `gorm:"size:255;comment:游戏订单号"`                        // 游戏订单号
	GyyxOrderNo   string    `gorm:"size:255;comment:平台订单号"`                        // 平台订单号
	Result        string    `gorm:"size:50;comment:支付结果"`                          // 支付结果
	ResultMessage string    `gorm:"size:255;comment:支付结果信息"`                       // 支付结果信息
	RmbYuan       float64   `gorm:"type:decimal(10,2);not null;comment:金额，保留两位小数"` // 金额
	ServerFlag    string    `gorm:"size:100;comment:服务器标识"`                        // 服务器标识
	CommonParam   string    `gorm:"size:255;comment:通用参数"`                         // 通用参数
	Timestamp     string    `gorm:"size:50;comment:时间戳"`                           // 时间戳
	CreatedAt     time.Time `gorm:"autoCreateTime;comment:创建时间"`                   // 创建时间
}

// InitDB 初始化数据库连接
func InitDB() error {
	dsn := config.GetDBConnectionString()
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&GameGoods{},
		&GameOrder{},
		&GameOrderPay{},
	)

	return err
}

func InsertOrder(order any) error {
	return DB.Create(order).Error
}

func GetOrderByGoodsId(goodsId int) (*GameGoods, error) {
	var goods GameGoods
	result := DB.Where("id = ?", goodsId).Order("created_at desc").First(&goods)
	return &goods, result.Error
}

func GetOrderByGoodsIitem(itemId string, singlePrice float64) (*GameGoods, error) {
	var goods GameGoods
	result := DB.Where("item = ? AND single_pric = ?", itemId, singlePrice).Order("created_at desc").First(&goods)
	return &goods, result.Error
}

func GetOrderByNo(orderNo string) (bool, error) {
	var exists bool
	result := DB.Model(&GameOrder{}).Select("1").Where("`order` = ? AND order_status = 0", orderNo).Limit(1).Find(&exists)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 查询你此订单号是否是未支付的状态
func GetOrderByNoPrice(orderNo string, single_price float64) (*GameOrder, error) {
	var order GameOrder
	result := DB.Where("`order` = ? AND single_price = ? AND order_status = 0", orderNo, single_price).First(&order)
	return &order, result.Error
}

// 查询是否存在未支付的订单
func GetOrderByUserAndItem(userId string, item string, singlePrice float64) (*GameOrder, error) {
	var order GameOrder
	// 仅当订单存在且未支付时才返回
	result := DB.Where("user_id = ? AND item = ? AND single_price = ? AND order_status = 0", userId, item, singlePrice).First(&order)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 订单不存在，返回nil
			return nil, nil
		}
		// 记录其他错误
		return nil, result.Error
	}
	return &order, nil
}

// 更改订单为已支付的状态
func UpdateOrderStatus(userId, orderNo string) error {
	if err := DB.Model(&GameOrder{}).Where("user_id = ? AND `order` = ?", userId, orderNo).Updates(map[string]interface{}{
		"order_status": 2,
	}).Error; err != nil {
		return err
	}
	return nil
}

// 更改订单为取消的状态
func UpdateOrderStatusCancel(userId, orderNo string) error {
	now := time.Now()
	if err := DB.Model(&GameOrder{}).Where("user_id = ? AND `order` = ?", userId, orderNo).Updates(map[string]interface{}{
		"order_status": 1,
		"deleted_at":   &now,
	}).Error; err != nil {
		return err
	}
	return nil
}

func GetOrderPayExists(gameOrderNo string, singlePric float64) (bool, error) {
	var exists bool
	result := DB.Model(&GameOrderPay{}).Select("1").Where("game_order_no = ? AND rmb_yuan = ?", gameOrderNo, singlePric).Limit(1).Find(&exists)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func GetOrderPayExistsByOrderNo(gameOrderNo string) (bool, error) {
	var exists bool
	result := DB.Model(&GameOrderPay{}).Select("1").Where("game_order_no = ?", gameOrderNo).Limit(1).Find(&exists)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func GetOrderPayByUserId(userId string, item string) (*GameOrderPay, error) {
	var orderPay GameOrderPay
	result := DB.Where("user_id = ? AND item = ? AND game_order_no != '' ", userId, item).First(&orderPay)
	return &orderPay, result.Error
}
