package dingding

//统计请假频率接口
type CountFrequencyLeave interface {
	SendFrequencyLeave(startWeek int) error                                      //实现发送消息功能
	CountFrequencyLeave(startWeek int, result map[string][]DingAttendance) error //实现统计次数到数据库功能
}

type CountFrequencyLate interface {
	SendFrequencyLate(startWeek int) error                                      //实现发送消息功能
	CountFrequencyLate(startWeek int, result map[string][]DingAttendance) error //实现统计次数到数据库功能
}
