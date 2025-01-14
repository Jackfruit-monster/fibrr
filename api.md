# 支付系统接口文档

## 基础信息

- 基础URL: `https://ex.xxxxx.cn`
- 接口前缀: `/api/pay`

## 通用响应格式

所有接口都会返回以下格式的响应：

```json
{
  "result": "success",    // 响应结果：success 表示成功
  "state": "",            // 状态信息
  "trace_id": "uuid",     // 请求追踪ID
  "data": {}              // 具体的响应数据（部分接口可能没有）
}
```


| 参数名 |   类型   | 必填 | 描述 |
|:------:|:------:|:--:|:----:|
| result | string | -  | 响应结果：success 表示成功 |
| state | string | -  | 状态信息 |
| trace_id | string | -  | 请求追踪ID |
| data |  json  | -  | 具体的响应数据（部分接口可能没有） |

## 接口列表

### 1. 商品查询接口

- **接口路径**: `/api/pay/goods`
- **请求方式**: GET

#### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 是 | 商品ID |

#### 请求示例

```
GET /api/pay/goods?id=1
```

#### 响应示例

```json
{
  "result": "success",
  "state": "",
  "trace_id": "9e65360f-4170-4519-9131-f535c007fbf8",
  "data": {
    "id": 1,
    "item": "商品属性",
    "single_pric": 99.99
  }
}
```


### 2. 创建订单接口

- **接口路径**: `/api/pay/create-order`
- **请求方式**: POST
- **Content-Type**: application/json

#### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|----|------|
| user_id | string | 是  | 用户ID |
| item | string | 是  | 商品属性 |
| item_id | string | 否  | 商品ID |
| single_pric | number | 是  | 单价 |

#### 请求示例

```json
{
  "user_id": "888888",
  "item": "商品属性",
  "item_id": "商品ID",
  "single_pric": 120.88
}
```

#### 响应示例

```json
{
  "result": "success",
  "state": "",
  "trace_id": "79ebbfca-da10-4aac-9cc8-98a590d842bc",
  "data": {
    "item": "商品属性",
    "item_id": 1,
    "order": "811-107552442731859968",
    "single_pric": 120.88,
    "user_id": "888888"
  }
}
```

### 3. 支付回调接口(飞猪回调，不需要对接)

- **接口路径**: `/api/pay/callback`
- **请求方式**: POST
- **Content-Type**: application/json

#### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| game_order_no | string | 是 | 游戏订单号 |
| gyyx_order_no | string | 是 | 系统订单号 |
| result | string | 是 | 结果 |
| result_message | string | 是 | 结果信息 |
| rmb_yuan | number | 是 | 支付金额（元） |
| server_flag | string | 是 | 服务器标识 |
| common_param | string | 是 | 通用参数 |
| timestamp | string | 是 | 时间戳 |
| sign | string | 是 | 签名 |
| signType | string | 是 | 签名类型 |

#### 请求示例

```json
{
  "game_order_no": "1234",
  "gyyx_order_no": "12345567",
  "result": "12345567",
  "result_message": "12345567",
  "rmb_yuan": 128,
  "server_flag": "12345567",
  "common_param": "12345567",
  "timestamp": "12345567",
  "sign": "12345567",
  "signType": "12345567"
}
```

#### 响应示例

```json
{
  "result": "success",
  "state": "",
  "trace_id": "8e4f079d-73e7-46a6-b945-26fd0d692763"
}
```

### 4. 验证订单接口

- **接口路径**: `/api/pay/verification`
- **请求方式**: POST
- **Content-Type**: application/json

#### 请求参数

| 参数名 | 类型 | 必填 |  描述  |
|:------:|:----:|:--:|:----:|
| user_id | string | 是  | 用户ID |
| item | string | 是  | 商品属性 |
| item_id | string | 否  | 商品ID |
| order | string | 否  | 订单编号 |

#### 请求示例

```json
{   
  "user_id" :"888888", 
  "item" :"商品属性",  
  "item_id" :"商品ID",  
  "order": "811-123456789"
}
```

#### 响应示例
```json
{
  "result": "success",
  "state": "",
  "trace_id": "8e4f079d-73e7-46a6-b945-26fd0d692763"
}
```

### 5. 取消订单接口

- **接口路径**: `/api/pay/cancel-order`
- **请求方式**: POST
- **Content-Type**: application/json

#### 请求参数

| 参数名 | 类型 | 必填 | 描述   |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |
| order | string | 是 | 订单编号 |

#### 请求示例

```json
{
  "user_id" :"13758666",
  "order" :"811-110424410053283840"
}
```

#### 响应示例
```json
{
  "result": "success",
  "state": "",
  "trace_id": "8e4f079d-73e7-46a6-b945-26fd0d692763"
}
```

## 错误码说明

> 注：具体错误码需要根据state字段进行说明，result只有success、fail

|   错误码   | 描述 | 解决方案 |
|:-------:|----|:--------:|
| success | 成功 | - |
|  fail   | 失败 | - |