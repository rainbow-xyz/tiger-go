package v1

import (
	"github.com/gin-gonic/gin"
	"saas_service/internal/pkg/domain/passport"
	dtoV1 "saas_service/internal/pkg/dto/passport/v1"
	"saas_service/internal/pkg/middlewares/inoauth"
	"saas_service/pkg/core"
)

type WalletHandler struct {
	WalletUsecase passport.WalletUsecase
}

func NewSaasHandler(g *gin.Engine, wa passport.WalletUsecase) {
	handler := &WalletHandler{
		WalletUsecase: wa,
	}

	//  内部服务使用的结构
	apiv1Inner := g.Group("/_inner/v1")
	apiv1Inner.Use(inoauth.InnerServiceCheckAuth())
	{

		ir := apiv1Inner.Group("/saas")
		{
			ir.GET("/walletgetbycond", handler.WalletGetByCond)
			ir.GET("/walletgetsinglebycond", handler.WalletGetSingleByCond)
			ir.GET("/walletgetsinglebyid", handler.WalletGetByID)
			ir.POST("/walleteditbyidex", handler.WalletEditByIDEx)
		}
	}
}

func (s *WalletHandler) WalletGetByCond(ctx *gin.Context) {
	var req dtoV1.WalletGetByCond
	ctx.ShouldBind(&req)
	wallet, err := s.WalletUsecase.WalletGetByCond(ctx, req.Fields, req.Conds)
	core.WriteResponseX(ctx, err, wallet)
}
func (s *WalletHandler) WalletGetSingleByCond(ctx *gin.Context) {
	var req dtoV1.WalletGetSingleByCond
	ctx.ShouldBind(&req)
	wallet, err := s.WalletUsecase.WalletGetSingleByCond(ctx, req.Fields, req.Conds)
	core.WriteResponseX(ctx, err, wallet)
}

func (s *WalletHandler) WalletGetByID(ctx *gin.Context) {
	var req dtoV1.WalletGetByID
	ctx.ShouldBind(&req)
	wallet, err := s.WalletUsecase.WalletGetByID(ctx, req.ID)
	core.WriteResponseX(ctx, err, wallet)
}

func (s *WalletHandler) WalletEditByIDEx(ctx *gin.Context) {
	var req dtoV1.WalletEditByIDEx
	ctx.ShouldBind(&req)
	wallet := passport.Wallet{ID: req.ID, Balance: req.Balance}
	rows, err := s.WalletUsecase.WalletEditByIDEx(ctx, wallet)
	core.WriteResponseX(ctx, err, rows)
}
