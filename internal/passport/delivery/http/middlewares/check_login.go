package middlewares

import (
	"github.com/gin-gonic/gin"
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"
)

func CheckLogin(userUcase passport.UserUsecase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := core.GetMetaReqParamsFromReqCtx(ctx)
		user, err := userUcase.CheckBUserLogin(ctx, req.AccessToken)
		if err != nil {
			core.WriteResponse(ctx, err, nil)
			ctx.Abort()
			return
		}

		ctx.Set("__user_id", user.ID)
		ctx.Set("__user_name", user.Name)
		ctx.Set("__user_phone", user.Phone)
		ctx.Set("__user_unionid", user.Unionid)
		ctx.Set("__user_headimgurl", user.Headimgurl)

		// 写入原生上下文
		//ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "__user_id", user.ID))
		//ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "__user_name", user.Name))
		//ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "__user_phone", user.Phone))
		//ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "__user_unionid", user.Unionid))
		//ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "__user_headimgurl", user.Headimgurl))

		ctx.Next()
	}
}
