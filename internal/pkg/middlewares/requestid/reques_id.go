package requestid

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func SetMyGinRequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := requestid.Get(ctx)
		ctx.Set("__x_request_id", requestID)
		ctx.Next()
	}
}
