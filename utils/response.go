// utils/response.go
package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ResponseCode 定义响应状态码
type ResponseCode int

const (
	SUCCESS ResponseCode = 200
	FAIL    ResponseCode = 400
)

// ResponseResult 定义响应结果类型
type ResponseResult string

const (
	ResultSuccess ResponseResult = "success"
	ResultFail    ResponseResult = "fail"
)

// Response 标准响应结构
type Response struct {
	Result    string      `json:"result"`
	State     string      `json:"state"`
	TraceID   string      `json:"trace_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
}

// Response 标准响应结构
type ResponseTiktok struct {
	ErrNo   int         `json:"err_no"`
	ErrTips string      `json:"err_tips"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// ResponseWrapper 响应包装器结构
type ResponseWrapper struct {
	ctx *fiber.Ctx
}

// NewResponse 创建新的响应包装器
func NewResponse(c *fiber.Ctx) *ResponseWrapper {
	return &ResponseWrapper{ctx: c}
}

// Success 返回成功响应
func (w *ResponseWrapper) Success() error {
	traceID := w.getTraceID()
	return w.ctx.Status(int(SUCCESS)).JSON(Response{
		Result:  string(ResultSuccess),
		State:   "",
		TraceID: traceID,
	})
}

func (w *ResponseWrapper) SuccessTiktok() error {
	traceID := w.getTraceID()
	return w.ctx.Status(int(SUCCESS)).JSON(ResponseTiktok{
		ErrNo:   0,
		TraceID: traceID,
		ErrTips: "success",
	})
}

// SuccessWithData 返回带数据的成功响应
func (w *ResponseWrapper) SuccessWithData(data interface{}) error {
	traceID := w.getTraceID()
	return w.ctx.Status(int(SUCCESS)).JSON(Response{
		Result:  string(ResultSuccess),
		State:   "",
		TraceID: traceID,
		Data:    data,
	})
}

// Fail 返回失败响应
func (w *ResponseWrapper) Fail(status int, state string) error {
	traceID := w.getTraceID()
	return w.ctx.Status(status).JSON(Response{
		Result:  string(ResultFail),
		State:   state,
		TraceID: traceID,
	})
}

// FailWithCode 返回带错误码的失败响应
func (w *ResponseWrapper) FailWithCode(status int, state string, errorCode string) error {
	traceID := w.getTraceID()
	return w.ctx.Status(status).JSON(Response{
		Result:    string(ResultFail),
		State:     state,
		TraceID:   traceID,
		ErrorCode: errorCode,
	})
}

// Custom 返回自定义响应
func (w *ResponseWrapper) Custom(status int, result string, state string) error {
	traceID := w.getTraceID()
	return w.ctx.Status(status).JSON(Response{
		Result:  result,
		State:   state,
		TraceID: traceID,
	})
}

// CustomWithData 返回带数据的自定义响应
func (w *ResponseWrapper) CustomWithData(status int, result string, state string, data interface{}) error {
	traceID := w.getTraceID()
	return w.ctx.Status(status).JSON(Response{
		Result:  result,
		State:   state,
		TraceID: traceID,
		Data:    data,
	})
}

// getTraceID 获取 trace-id
func (w *ResponseWrapper) getTraceID() string {
	if traceID := w.ctx.Locals("trace_id"); traceID != nil {
		return traceID.(string)
	}
	return ""
}
