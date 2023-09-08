package params

type ParamGetTasks struct {
	RobotId string `json:"robot_id" binding:"required"`
}
type ParamRemoveTask struct {
	TaskId string `json:"task_id" binding:"required"`
}
type ParamStopTask struct {
	TaskId string `json:"task_idd" binding:"required"`
}
type ParamReStartTask struct {
	ID uint `json:"id" binding:"required"`
}
