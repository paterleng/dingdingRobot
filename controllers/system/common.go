package system

import (
	"ding/response"
	"github.com/gin-gonic/gin"
)

func WelcomeHandler(c *gin.Context) {
	response.OkWithMessage("hello", c)
}
