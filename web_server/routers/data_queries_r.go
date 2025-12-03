package routers

import (
	e "web_server/environments"
	v "web_server/validations"
	c "web_server/controllers"

	"github.com/gin-gonic/gin"
)

func DataQueriesRouter(g *gin.RouterGroup) {
	g.GET(e.SearchDataPath, v.CheckSearchData, c.SearchData)
}
