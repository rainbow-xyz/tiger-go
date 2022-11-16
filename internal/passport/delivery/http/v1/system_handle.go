package v1

import (
	"github.com/gin-gonic/gin"
	"saas_service/pkg/core"
)

type SystemHandler struct {
}

func NewSystemHandler(g *gin.Engine) {
	handler := &SystemHandler{}

	sys := g.Group("/system")
	{
		sys.GET("/healthcheck", handler.HealthCheck)
	}
}

func (s *SystemHandler) HealthCheck(ctx *gin.Context) {
	core.WriteResponseX(ctx, nil, "SUCCESS!")
}
