package dingding

import (
	ding2 "ding/controllers/ding"
	"ding/global"
	"ding/model/dingding"
	"fmt"

	"github.com/gin-gonic/gin"
)

func SetupDing(System *gin.RouterGroup) {
	Dept := System.Group("/dept")
	{
		Dept.GET("/ImportDeptData", ding2.ImportDeptData)                       // 递归获取部门列表存储到数据库
		Dept.GET("/getSubDepartmentListById", ding2.GetSubDepartmentListByID)   // 官方接口获取子部门
		Dept.GET("/getSubDepartmentListById2", ding2.GetSubDepartmentListByID2) // 从数据库中一层一层的取出部门
		Dept.GET("/getDeptListFromMysql", ding2.GetDeptListFromMysql)           //从数据库中取出部门信息，包括该部门的负责人
		Dept.PUT("/updateDept", ding2.UpdateDept)                               // 更新部门信息，用来设置机器人token，各种开关
		Dept.PUT("/updateSchool", ding2.UpdateSchool)                           //更新部门是否在校信息
		Dept.PUT("/setDeptManager", ding2.SetDeptManager)                       //更新部门负责人
		Dept.GET("/getUserByDeptid", ding2.GetUserByDeptId)                     //根据部门id查询用户信息
	}

	AttendanceGroup := System.Group("/attendanceGroup")
	{
		AttendanceGroup.GET("/ImportAttendanceGroupData", ding2.ImportAttendanceGroupData)    //将考勤组信息导入到数据库中
		AttendanceGroup.PUT("/updateAttendanceGroup", ding2.UpdateAttendanceGroup)            //考勤组开关
		AttendanceGroup.GET("/GetAttendanceGroupList", ding2.GetAttendanceGroupListFromMysql) //批量获取考勤组
	}
	LeaveGroup := System.Group("/leave")
	{
		LeaveGroup.POST("/SubscribeToSomeone", ding2.SubscribeToSomeone) //订阅某人考勤情况
		LeaveGroup.DELETE("/Unsubscribe", ding2.Unsubscribe)             //取消订阅
	}
	User := System.Group("/user")
	{
		User.GET("/getUserInfo", ding2.GetUserInfo)
		User.POST("ImportDingUserData", ding2.ImportDingUserData)  //将钉钉用户导入到数据库中
		User.POST("/UpdateDingUserAddr", ding2.UpdateDingUserAddr) // 更新用户的博客和简书地址
		User.GET("/GetAllUsers", ding2.SelectAllUsers)             // 查询所有用户信息
		User.GET("/GetAllJinAndBlog", ding2.FindAllJinAndBlog)
		User.GET("/showQRCode", func(c *gin.Context) {
			username, _ := c.Get(global.CtxUserNameKey)
			c.File(fmt.Sprintf("Screenshot_%s.png", username))
		})
		User.GET("getQRCode", ding2.GetQRCode)                                  //获取群聊基本信息已经群成员id
		User.GET("/getAllTask", ding2.GetAllTask)                               //获取所有定时任务，包括暂停的任务
		User.GET("/getActiveTask", ding2.GetAllActiveTask)                      //查看所有的活跃任务,也就是手动更新，后续可以加入casbin，然后就是管理员权限
		User.POST("/MakeupSign", ding2.MakeupSign)                              //为用户补签到并返回用户联系签到次数
		User.GET("/getWeekConsecutiveSignNum", ding2.GetWeekConsecutiveSignNum) //获取用户当周连续签到次数
		User.GET("/getWeekSignNum", ding2.GetWeekSignNum)                       //根据第几星期获取用户签到次数（使用redis的bitCount函数）
		User.GET("/getWeekSignDetail", ding2.GetWeekSignDetail)                 //获取用户某个星期签到情况，默认是当前所处的星期，构建成为一个有序的HashMap
		User.GET("/getDeptIdByUserId", ding2.GetDeptByUserId)                   //通过userid查询部门id
		User.POST("/resetWeek", dingding.ResetWeek)                             //重置维护的周数
	}
	Robot := System.Group("robot")
	{
		Robot.POST("/pingRobot", ding2.PingRobot)
		Robot.POST("/addRobot", ding2.AddRobot)
		Robot.DELETE("/removeRobot", ding2.RemoveRobot)
		Robot.PUT("/updateRobot", ding2.AddRobot) //更新机器人直接使用
		Robot.GET("/getSharedRobot", ding2.GetSharedRobot)
		Robot.GET("/getRobotDetailByRobotId", ding2.GetRobotDetailByRobotId)
		Robot.GET("/getRobotBaseList", ding2.GetRobots)            //获取所有机器人
		Robot.POST("/cronTask", ding2.CronTask)                    //发送定时任务
		Robot.POST("/getTaskList", ding2.GetTaskList)              //加载定时任务
		Robot.POST("/stopTask", ding2.StopTask)                    //暂停定时任务
		Robot.DELETE("/removeTask", ding2.RemoveTask)              //移除定时任务
		Robot.POST("/reStartTask", ding2.ReStartTask)              //重启定时任务
		Robot.PUT("/editTaskContent", ding2.EditTaskContent)       //编辑定时任务的内容
		Robot.GET("/getTaskDetail", ding2.GetTaskDetail)           //获取定时任务详情
		Robot.GET("/getAllPublicRobot", ding2.GetAllPublicRobot)   //获取所有的公共机器人
		Robot.PUT("/alterResultByRobot", ding2.AlterResultByRobot) //修改部门考勤果推送到哪个群,给我一个要修改到哪个群的公共机器人的token，你要修改的部门id
		Robot.GET("/updateMobile", ding2.UpdateMobile)
		Robot.POST("/singleChat", ding2.SingleChat)
	}
	//机器人问答模块
	QuAndAn := System.Group("/quAndAn")
	{
		QuAndAn.POST("/updateData", ding2.UpdateData)   //上传资源
		QuAndAn.DELETE("/deleteData", ding2.DeleteData) //删除资源
		QuAndAn.PUT("/putData", ding2.PutData)          //修改资源
		QuAndAn.POST("/getData", ding2.GetData)         //查询资源
	}
}
