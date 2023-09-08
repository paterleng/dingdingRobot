package ding

type ParamUpdateUpdateAttendanceGroup struct {
	GroupId           int  `gorm:"primaryKey" json:"group_id"` //考勤组id
	IsRobotAttendance bool `json:"is_robot_attendance"`        //该考勤组是否开启机器人查考勤 （相当于是总开关）
	IsSendFirstPerson int  `json:"is_send_first_person"`       //该考勤组是否开启推送每个部门第一位打卡人员 （总开关）
	IsReady           int  `json:"is_ready"`                   //是否开启预备考勤
	ReadyTime         int  `json:"ready_time"`                 //提 前几分钟开启预备考勤
}
type ParamGetAttendGroup struct {
}
