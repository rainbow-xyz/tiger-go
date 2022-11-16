package core

import (
	"github.com/gin-gonic/gin"
)

type MetaReqParams struct {
	OS          string `json:"os,omitempty" form:"os"`
	Version     string `json:"version,omitempty" form:"version"`
	AccessToken string `json:"access_token,omitempty" uri:"access_token" form:"access_token" gorm:"column:access_token"`
	XToken      string `json:"x_token,omitempty" uri:"x_token" form:"x_token" gorm:"column:x_token"`
}

var meteReqParams MetaReqParams

func GetMetaReqParamsFromReqCtx(ctx *gin.Context) MetaReqParams {
	meteReqParams.OS = ctx.GetHeader("X-OS")
	meteReqParams.Version = ctx.GetHeader("X-Version")
	meteReqParams.AccessToken = ctx.GetHeader("X-Access-Token")
	meteReqParams.XToken = ctx.GetHeader("X-Token")

	return meteReqParams
}
