package model

import (
	"github.com/gin-gonic/gin"
)

type Jk struct {
	Names []string `json:"names"`
}

func JkFunc(c *gin.Context, p *Jk) (err error) {
	//UserIds := []string{}
	//for _, name := range p.Names {
	//	userid := ""
	//	err = global.GLOAB_DB.Model(&dingding.Tele{}).Select("user_id").Where("personname = ?", name).First(&userid).Error
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	UserIds = append(UserIds, userid)
	//}
	//AtUserIds := make([]common.AtUserId, len(UserIds))
	//for i := 0; i < len(AtUserIds); i++ {
	//	AtUserIds[i].AtUserId = UserIds[i]
	//}
	//at := common.At{
	//	AtUserIds: AtUserIds,
	//	IsAtAll:   false,
	//}
	//text := common.MsgText{
	//	At:      at,
	//	Msgtype: "text",
	//	Text: common.Text{
	//		Content: "未在规定时间内阅读《产品经理必须要懂的那些事》完成",
	//	},
	//}
	//Jksend := params.ParamSend{
	//	Version:     "v2",
	//	MsgText:     text,
	//	RobotId:     utils.JkRobotId,
	//	RepeateTime: "立即发送",
	//}
	//err, _ = Send(c, &Jksend)
	//if err != nil {
	//	zap.L().Error("京科接口推送消息失败", zap.Error(err))
	//}
	//return
	return
}
