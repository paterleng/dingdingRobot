package gxp

import (
	"ding/global"
	"ding/model/common"
	"ding/model/dingding"
	"ding/utils"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	_ "strconv"
	"strings"
	"time"
)

//const RobotToken = "11e07612181c7b596e49e80d26cb368318a2662c0f6affd453ccfd3d906c2431"

func getMysqlToken() (token string) {
	err := global.GLOAB_DB1.Table("configs").Where("k = ?", "token").Select("v").Scan(&token).Error
	if err != nil {
		zap.L().Error("通过mysql查询机器人token错误", zap.Error(err))
	}
	return
}
func CronSendOne() (err error) {
	spec := fmt.Sprintf("0 %v %v ? * * ", utils.StartMin, utils.StartHour)

	//开启定时器，定时每晚10：00(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		message := "大家现在到宿舍的话就可以开始报备了[爱意]"
		fmt.Println(message)
		zap.L().Info("message编辑完成，开始封装发送信息参数")
		p := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				At: common.At{
					IsAtAll: true,
				},
				Text: common.Text{
					Content: message,
				},
				Msgtype: "text",
			},
			RobotId: getMysqlToken(),
		}
		err := (&dingding.DingRobot{
			RobotId: getMysqlToken(),
		}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：00定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：00定时任务", entryID)
	return
}
func CronSendTwo() (err error) {
	spec := fmt.Sprintf("0 %v %v ? * * ", utils.RemindMin, utils.RemindHour)
	//开启定时器，定时22：20提醒未到宿舍人员(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		day := time.Now().Format("2006-01-02")
		var AllUsers []dingding.TongXinUser
		var atRobotUsers []dingding.TongXinUser
		var notAtRobotUserIds []common.AtUserId
		global.GLOAB_DB1.Where("is_school = ?", 1).Preload("Records", "created_at like ?", "%"+day+"%").Find(&AllUsers)
		for _, user := range AllUsers {
			if user.Records == nil || len(user.Records) == 0 {
				notAtRobotUserIds = append(notAtRobotUserIds, common.AtUserId{
					AtUserId: user.ID,
				})
			} else {
				atRobotUsers = append(atRobotUsers, user)
			}
		}
		message := "还有十分钟就结束报备了，还没报备的同学抓紧时间了[吃瓜]\n"
		if len(notAtRobotUserIds) == 0 && len(atRobotUsers) == len(AllUsers) {
			message = "截至目前所有人员已报备，谢谢大家配合[送花花][送花花]"
		} else {
			message = "截至目前还有以下同学未报备是否到达宿舍："
		}

		zap.L().Info("message编辑完成，开始封装发送信息参数")
		p := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				At: common.At{
					AtUserIds: notAtRobotUserIds,
				},
				Text: common.Text{
					Content: message,
				},
				Msgtype: "text",
			},
			RobotId: getMysqlToken(),
		}
		err := (&dingding.DingRobot{
			RobotId: getMysqlToken(),
		}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：20定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：20定时任务", entryID)
	return
}

// CronSendThree 晚上10：35统计结果发给gxp
func CronSendThree() (err error) {
	spec := fmt.Sprintf("0 %v %v ? * * ", utils.EndMin, utils.EndHour)
	//开启定时器，定时22：30发送私聊以及群消息
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		day := time.Now().Format("2006-01-02")
		var AllUsers []dingding.TongXinUser
		var atRobotUsers []dingding.TongXinUser
		var notAtRobotUsers []dingding.TongXinUser
		global.GLOAB_DB1.Where("is_school = ?", 1).Preload("Records", "created_at like ?", "%"+day+"%").Find(&AllUsers)
		for _, user := range AllUsers {
			if user.Records == nil || len(user.Records) == 0 {
				notAtRobotUsers = append(notAtRobotUsers, user)
			} else {
				atRobotUsers = append(atRobotUsers, user)
			}
		}
		message := "应留校" + strconv.Itoa(len(AllUsers)) + "人\n已到寝"
		var atRoomNum int
		var atRoomUsers []dingding.TongXinUser
		var notAtRoomUsers []dingding.TongXinUser
		for _, atRobotUser := range atRobotUsers {
			if strings.Contains(atRobotUser.Records[len(atRobotUser.Records)-1].Content, "宿舍") || strings.Contains(atRobotUser.Records[len(atRobotUser.Records)-1].Content, "寝室") {
				atRoomUsers = append(atRoomUsers, atRobotUser)
				atRoomNum++
			} else {
				notAtRoomUsers = append(notAtRoomUsers, atRobotUser)
			}
		}
		//var numbers = len(atRobotUsers)
		message += strconv.Itoa(atRoomNum) + "人\n"
		//for _, atRobotUser := range atRobotUsers {
		//	message += atRobotUser.Name + " "
		//}
		message += "特殊原因：\n"
		for _, notAtRoomUser := range notAtRoomUsers {
			message += notAtRoomUser.Name + ":" + notAtRoomUser.Records[len(notAtRoomUser.Records)-1].Content + "\n"
		}
		message += "\n未报备：\n"
		for _, notAtRobotUser := range notAtRobotUsers {
			message += notAtRobotUser.Name + "  "
		}
		message += "\n\n已到寝" + strconv.Itoa(atRoomNum) + "人："
		for _, user := range atRoomUsers {
			message += user.Name + " "
		}

		zap.L().Info("message编辑完成，开始封装发送信息参数")
		//关鑫鹏个人的userid
		//var userId = []string{"01144160064621256183"}
		//闫佳鹏的userid
		var userId = []string{"413550622937553255", "01144160064621256183"}
		//私聊消息
		p := &dingding.ParamChat{
			RobotCode: "dingpi0onbdchuv5anhn",
			UserIds:   userId,
			MsgKey:    "sampleText",
			MsgParam:  message,
		}
		err := (&dingding.DingRobot{
			RobotId: getMysqlToken(),
		}).GxpSingleChat(p)
		//发在群里的提醒
		p2 := &dingding.ParamCronTask{
			MsgText: &common.MsgText{
				Text: common.Text{
					Content: "今天的报备结束，谢谢大家配合[送花花]",
				},
				Msgtype: "text",
			},
			RobotId: getMysqlToken(),
		}
		err = (&dingding.DingRobot{
			RobotId: getMysqlToken(),
		}).SendMessage(p2)

		//p := &dingding.ParamCronTask{
		//	MsgText: &common.MsgText{
		//		At: common.At{
		//			AtUserIds: notAtRobotUserIds,
		//		},
		//		Text: common.Text{
		//			Content: message,
		//		},
		//		Msgtype: "text",
		//	},
		//	RobotId: RobotToken,
		//}
		//err := (&dingding.DingRobot{
		//	RobotId: RobotToken,
		//}).SendMessage(p)
		if err != nil {
			zap.L().Error("发送关鑫鹏22：35定时任务失败", zap.Error(err))
			return
		}
	})
	fmt.Println("关鑫鹏22：35定时任务", entryID)
	return
}
