package user

import (
	"github.com/gin-gonic/gin"
	"saas_service/pkg/core"
)

type GetUserByPhoneReq struct {
	core.MetaReqParams
	Phone string `json:"phone" form:"phone" uri:"phone" validate:"required,number,len=11"`
}

func (u *Handler) GetUserByPhone(ctx *gin.Context) {

	var req GetUserByPhoneReq
	ctx.ShouldBind(&req)

	err := core.ValidateStruct(req)
	if err != nil {
		core.WriteResponseX(ctx, err, nil)
		return
	}

	user, err := u.UUsecase.GetUserByPhone(ctx, req.Phone)
	core.WriteResponseX(ctx, err, user)
}
