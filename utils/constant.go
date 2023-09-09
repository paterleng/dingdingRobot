package utils

const (
	ConfigFile = "./conf/config.yaml"
	//ConfigFile = "./conf/config.yaml"
	Version                   = "v2"
	CropId                    = "CropId"
	AccessToken               = "AccessToken"
	GxpAccessToken            = "GxpAccessToken"
	ConstTypingInvitationCode = "ConstTypingInvitationCode"
	AppKey                    = "AppKey"
	GxpAppKey                 = "GxpAppKey"
	AppSecret                 = "AppSecret"
	GxpAppSecret              = "GxpAppSecret"
	AttendanceSucc            = "考勤已记录~"
	AttendanceUpdateSucc      = "考勤已成功更新~"
	AttendanceFail            = "考勤记录失败，请联系技术人员解决欧~"
	TypingInviationFail       = "打字邀请码获取失败，请联系技术人员解决欧~"
	TypingInviationSucc       = "打字邀请码获取成功~"
	JkRobotId                 = "1317ac8ee5004f475046029a3f1bb94873a7dd46897e6845f318a71d0402a1ea"
	TestRobotToken            = "7e07aeb5a804631802f0347cbb98b579f5a0fbf9883da891b9188555efd42d97"
	Delay                     = '5'  //考勤向后延长时间
	Advance                   = "10" //提前多久提醒打卡,提前时间加上延后时间
	SanQiXiaoZhao             = "1317ac8ee5004f475046029a3f1bb94873a7dd46897e6845f318a71d0402a1ea"
	LeZhiSanQi                = "a7501dc76c0bf9b4afcb756a77053483726f4ddf5ef1d7ef1ae8ac8156931714"
	SanQiSheZhao              = ""
	LeZhiAllPeopleRobotName   = "2022寒假乐知常驻群聊机器人"
	SpecMorning               = "00 26 10 ? * 1-6" //周一到周六    标准用法0-6 or SUN-SAT
	SpecAfternoon             = "00 49 14 ? * 1-6"
	SpecEvening               = "0 30 19 ? * 1-6"
	TokenLength               = 32
	Morning                   = "早上"
	Afternoon                 = "下午"
	Evening                   = "晚上"
)
