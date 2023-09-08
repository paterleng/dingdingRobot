package personal

import (
	v1 "ding/controllers/personal"
	"github.com/gin-gonic/gin"
)

func SetupPersonal(Personal *gin.RouterGroup) {
	Personal.POST("/jk", v1.Jk)
	Personal.POST("/zjq", v1.Zjq)
	Personal.POST("/lxy", v1.Lxy)
}
