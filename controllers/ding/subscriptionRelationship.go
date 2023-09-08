package ding

import (
	dingding2 "ding/model/dingding"
	"ding/response"
	"github.com/gin-gonic/gin"
)

func SubscribeToSomeone(c *gin.Context) {
	relationship := &dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	err := relationship.SubscribeSomeone()
	if err != nil {
		response.FailWithMessage("订阅失败", c)
	} else {
		response.OkWithMessage("订阅成功", c)
	}
}

func Unsubscribe(c *gin.Context) {
	relationship := dingding2.SubscriptionRelationship{}
	relationship.Subscriber = c.Query("subscriber")
	relationship.Subscribee = c.Query("subscribee")
	err := relationship.UnsubscribeSomeone()
	if err != nil {
		response.FailWithMessage("取消订阅失败", c)
	} else {
		response.OkWithMessage("取消订阅成功", c)
	}
}
