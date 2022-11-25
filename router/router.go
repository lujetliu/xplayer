package router

import (
	"xplayer/service"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	route := gin.Default()
	route.GET("", service.ServerH264)
	return route
}
