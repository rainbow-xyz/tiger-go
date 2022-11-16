package inoauth

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/errors"
	"saas_service/internal/pkg/code"
	"saas_service/pkg/core"
	"saas_service/pkg/setting"
)

var xToken string

func NewXToken(config *setting.Config) {
	xToken = config.AppCfg.OAuthTmpToken
}

func InnerServiceCheckAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("X-Token")
		err := innerServiceCheckAuth(token)
		if err != nil {
			core.WriteResponse(ctx, err, nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// innerServiceCheckAuth sxy内部服务调用 先简单做个map来控制
func innerServiceCheckAuth(token string) error {
	if token == xToken {
		return nil
	}
	return errors.WithCode(code.ErrTokenInvalid, "无效的token")

}
