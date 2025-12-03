package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func MainRouter(router *gin.Engine) {
	env := c.GetWebServerEnv()
	config := cors.DefaultConfig()
	config.AllowOrigins = env.AllowOrigins
	config.AllowMethods = env.AllowMethods
	config.AllowHeaders = env.AllowHeaders
	
	config.AllowCredentials = true

	router.Use(cors.New(config))
	DataQueriesRouter(router.Group(e.DataQueriesBasePath))
}
