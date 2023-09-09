package global

import (
	"errors"
	"github.com/gin-gonic/gin"
)

var ErrorUserNotLogin = errors.New("用户登录状态有误，请重新登录")
var ErrorRobotNotLogin = errors.New("机器人未登录")
var ErrorCornTabNotGet = errors.New("定时任务获取失败")

const CtxUserIDKey = "user_id"
const CtxIDKey = "id"
const CtxUserNameKey = "userName"
const CtxUserAuthorityIDKey = "authority_id"

const CtxRobotIDKey = "robotID"
const CtxCornTab = "task"

// GetCurrentUser 获取当前登录用户的ID
func GetCurrentUserId(c *gin.Context) (UserID string, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	UserID = uid.(string) // 进行类型断言
	return UserID, err
}

// GetCurrentAuthorityIDKey 获取当前登录用户是否是负责人
//func GetCurrentAuthorityIDKey(c *gin.Context) (UserID string, err error) {
//	uid, ok := c.Get(CtxUserAuthorityIDKey)
//	if !ok {
//		err = ErrorUserNotLogin
//		return
//	}
//	UserID = uid.(string) // 进行类型断言
//	return UserID, err
//}

func GetCurrentUserName(c *gin.Context) (UserName string, err error) {
	uName, ok := c.Get(CtxUserNameKey)
	if !ok {
		err = errors.New("获取当前登录用户的姓名失败")
		return
	}
	UserName = uName.(string)
	return
}

//func GetCurrentDeptId(c *gin.Context) (DeptID string, err error) {
//	uName, ok := c.Get(CtxUserNameKey)
//	if !ok {
//		err = errors.New("获取当前登录用户的姓名失败")
//		return
//	}
//	UserName = uName.(string)
//	return
//}
