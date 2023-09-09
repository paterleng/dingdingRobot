package system

import (
	"ding/model/system"
	"github.com/gin-gonic/gin"
)

func GetMenu(c *gin.Context) {
	(&system.SysAuthorityMenu{}).GetMenu(c)
}
