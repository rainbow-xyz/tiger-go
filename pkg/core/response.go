package core

import (
	"net/http"
	"saas_service/internal/pkg/code"
	"saas_service/pkg/xlog"

	"github.com/gin-gonic/gin"
	"github.com/marmotedu/errors"
)

type ErrResponse struct {
	// Code 定义业务错误码.
	Code int `json:"code"`

	// Message 错误详细信息，注意能对外暴露的才能放在里面
	Message string `json:"message"`

	// Reference 错误引用文档提示如何解决
	Reference string `json:"reference,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type SuccessResponse struct {
	// Code 定义业务错误码.
	Code int `json:"code"`

	// Message 错误详细信息，注意能对外暴露的才能放在里面
	Message string `json:"message"`

	// Reference 错误引用文档提示如何解决
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id,omitempty"`
}

func WriteResponse(c *gin.Context, err error, data interface{}) {

	if err != nil {
		coder := errors.ParseCoder(err)
		c.Set("myErr", err)
		c.JSON(coder.HTTPStatus(), ErrResponse{
			Code:      coder.Code(),
			Message:   coder.String(),
			Reference: coder.Reference(),
		})

		return
	}

	c.JSON(http.StatusOK, data)
	return
}

func WriteResponseX(c *gin.Context, err error, data interface{}) {
	reqID, ok := c.Get("__x_request_id")
	if !ok {
		reqID = ""
	}
	if err != nil {
		coder := errors.ParseCoder(err)
		msg := coder.String()
		c.Set("myErr", err)

		if errors.IsCode(err, code.ErrValidationCustom) {
			// 特殊处理请求参数错误
			msg = "Invalid params (" + err.Error() + ")"
		}

		c.JSON(coder.HTTPStatus(), ErrResponse{
			Code:      coder.Code(),
			Message:   msg,
			Reference: coder.Reference(),
			RequestID: reqID.(string),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:      200,
		Message:   "Success",
		Data:      data,
		RequestID: reqID.(string),
	})
	return
}

// WriteResponseForZg 兼容之前项目返回参数
func WriteResponseForZg(c *gin.Context, err error, data interface{}) {
	if err != nil {
		xlog.XSErrorF(c, "%#+v", err)
		coder := errors.ParseCoder(err)
		c.JSON(coder.HTTPStatus(), gin.H{
			"errno": coder.Code(),
			"error": coder.String(),
			"data":  gin.H{},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errno": 200,
		"error": "Success",
		"data":  data,
	})
	return
}
