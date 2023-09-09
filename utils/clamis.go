package utils

//import (
//	"ding/initialize/jwt"
//	"github.com/gin-gonic/gin"
//	"go.uber.org/zap"
//)
//
//func GetClaims(c *gin.Context) (*jwt.MyClaims, error) {
//	//todo 需要修改
//	token := c.Request.Header.Get("x-token")
//	claims, err := (&jwt.MyClaims{}).ParseToken(token)
//	if err != nil {
//		zap.L().Error("从Gin的Context中获取从jwt解析信息失败, 请检查请求头是否存在x-token且claims是否为规定结构")
//	}
//	return claims, err
//}
//
//// GetUserAuthorityId 从Gin的Context中获取从jwt解析出来的用户角色id
//func GetUserAuthorityId(c *gin.Context) uint {
//	if claims, exists := c.Get("claims"); !exists {
//		if cl, err := GetClaims(c); err != nil {
//			return 0
//		} else {
//			return cl.AuthorityID
//		}
//	} else {
//		waitUse := claims.(*jwt.MyClaims)
//		return waitUse.AuthorityID
//	}
//}
//func GetUserID(c *gin.Context) string {
//	if claims, exists := c.Get("claims"); !exists {
//		if cl, err := GetClaims(c); err != nil {
//			return ""
//		} else {
//			return cl.UserId
//		}
//	} else {
//		waitUse := claims.(*jwt.MyClaims)
//		return waitUse.UserId
//	}
//}
//func GetUserInfo(c *gin.Context) *jwt.MyClaims {
//	if claims, exists := c.Get("claims"); !exists {
//		if cl, err := GetClaims(c); err != nil {
//			return nil
//		} else {
//			return cl
//		}
//	} else {
//		waitUse := claims.(*jwt.MyClaims)
//		return waitUse
//	}
//}
