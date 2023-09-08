package params

type ParamGetAttendances struct {
	ChatBotUserId string `json:"chat_bot_user_id"`
}

type ParamGetDeptFirstShowUpMorning struct {
	Frequency string `json:"frequency"` //推送消息的频率（默认是每天推送一次）
	GroupID   int    `json:"group_id"`  //考勤组id
	Token     string `json:"token"`
}
type ParamAllDepartDetail struct {
	DeptIds []int
	Token   string
}

//推送每个部门的考勤
type ParamAllDepartAttendByRobot struct {
	GroupId int `json:"group_id"`
}
