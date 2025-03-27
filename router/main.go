package router

import (
	"embed"
	"one-api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte) {
	router.Use(middleware.CORS())

	// Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 其他路由设置
	SetWebRouter(router, buildFS, indexPage)
	SetApiRouter(router)
	SetRelayRouter(router)
	SetDashboardRouter(router)
}
