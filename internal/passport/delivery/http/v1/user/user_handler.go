package user

import (
	"saas_service/internal/pkg/domain/passport"
	dtoV1 "saas_service/internal/pkg/dto/passport/v1"
	"saas_service/internal/pkg/middlewares/inoauth"
	"saas_service/pkg/core"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	UUsecase passport.UserUsecase
}

func NewUserHandler(g *gin.Engine, us passport.UserUsecase) {
	handler := &Handler{
		UUsecase: us,
	}

	//  内部服务使用的结构
	apiv1Inner := g.Group("/_inner/v1")
	apiv1Inner.Use(inoauth.InnerServiceCheckAuth())
	{
		userInner := apiv1Inner.Group("user")
		{
			userInner.GET("/get", handler.GetUserByID)
		}
	}
}

func (u *Handler) GetUserByID(ctx *gin.Context) {
	var req dtoV1.UserGetReq
	ctx.ShouldBind(&req)
	user, err := u.UUsecase.GetUserByID(ctx, req.UserID)
	core.WriteResponseX(ctx, err, dtoV1.NewUserGetByIDRespFromDomainUser(user))
}

func (u *Handler) HealthCheck(ctx *gin.Context) {
	core.WriteResponseX(ctx, nil, "SUCCESS!")
}
