package cron

import (
	"ding/global"
	dingding2 "ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
)

func Reboot() (err error) {
	global.GLOAB_CORN.Start() //在项目最初运行的时候，就要打开定时器，不然无法恢复成功
	//此处需要读取一下数据库中task表的内容，把task重新加载一遍，只去deleted_at为空的定时任务
	tasks := []dingding2.Task{}
	//链式预加载查询
	err = global.GLOAB_DB.Model(&tasks).Preload("MsgText.At.AtMobiles").
		Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").
		Where("deleted_at is null").
		Find(&tasks).Error //拿到所有的处在活跃状态的定时任务
	//a := "会议通知：\n时间：明天周四下午14：30\n地点：0#910\n主题：三期校招周会\n参会人员：三期校招全部成员、李老师、大学姐\n\n备注：该消息由机器人按时推送，如有变动，负责人请及时调整[送花花]"
	if err != nil {
		zap.L().Error("项目重启恢复定时查询数据库失败", zap.Error(err))
		return
	}
	tid := -1
	tasker := func() {}
	for _, task := range tasks {
		p := dingding2.ParamCronTask{
			MsgText:     task.MsgText,
			MsgLink:     task.MsgLink,
			MsgMarkDown: task.MsgMarkDown,
			RobotId:     task.RobotId,
		}
		d := dingding2.DingRobot{
			RobotId: task.RobotId,
			Secret:  task.Secret,
		}
		tasker = func() {
			err := d.SendMessage(&p)
			if err != nil {
				//zap.L().Error(fmt.Sprintf("恢复任务失败！发送人:%s,对应机器人:%s", username, robotname), zap.Error(err))
				return
			} else {
				//zap.L().Info(fmt.Sprintf("恢复任务成功！发送人:%s,对应机器人:%s", username, robotname))
			}
		}
		//	// 添加定时任务
		TaskID, err := global.GLOAB_CORN.AddFunc(task.Spec, tasker)
		if err != nil {
			//zap.L().Error("项目重启后恢复定时任务失败,失败原因：", zap.Error(err))
			//zap.L().Error(fmt.Sprintf("该任务所属人：%s,所属机器人：%s,"+
			//"人物名：%s,任务具体消息:%s,任务具体定时规则：%s", username, robotname, message, detailTimeForUser))
			return err
		}
		tid = int(TaskID)
		//oldId := task.TaskID
		err = global.GLOAB_DB.Model(&task).Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Where("deleted_at is null").Update("task_id", tid).Error
		if err != nil {
			//zap.L().Error("重启项目后更新任务id失败", zap.Error(err))
			return err
		}
		//zap.L().Info("项目重启后恢复定时任务成功")
		zap.L().Info(fmt.Sprintf("该任务所属人：%v,所属机器人：%v,"+
			"任务名：%s,任务具体消息:%s,任务具体定时规则：%s", task.UserName, task.RobotName, task.TaskName, task.MsgText.Text.Content, task.DetailTimeForUser))
	}

	global.GLOAB_CORN.Start() //在项目最初运行的时候，就要打开定时器，不然无法恢复成功
	return err
}
