package params

// ParamSignUp 定义请求的结构体参数
type ParamSignUp struct {
	Username   string `json:"username"  binding:"required" from:"username"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录时请求参数
type ParamLogin struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type ParamLoginByToken struct {
	Token string `json:"token"`
}
type ParamOutGoing struct {
}

type ParamSearchUser struct {
	RobotId    string `json:"robot_id" binding:"required"`
	PersonName string `json:"person_name" binding:"required"`
}
type ParamMakeupSign struct {
	Userid    string `json:"userid"`
	Year      int    `json:"year"`
	UpOrDown  int    `json:"up_or_down"`
	StartWeek int    `json:"start_week"`
	WeekDay   int    `json:"weekDay"`
	Diff      int    `json:"diff"` //1是正签，0是反签 暂时都默认成1
	MNE       int    `json:"mne"`  //早中晚
}
type ParamGetWeekConsecutiveSignNum struct {
	Userid    string `json:"userid"`
	Year      int    `json:"year"`
	UpOrDown  int    `json:"up_or_down"`
	StartWeek int    `json:"start_week"`
	WeekDay   int    `json:"week_day"`
	MNE       int    `json:"mne"` //早中晚
}
type ParamGetWeekSignNum struct {
	Userid    string `json:"userid"`
	Year      int    `json:"year"`
	UpOrDown  int    `json:"up_or_down"`
	StartWeek int    `json:"start_week"`
}
type ParamGetWeekSignDetail struct {
	Userid    string `json:"userid"`
	Year      int    `json:"year"`
	UpOrDown  int    `json:"up_or_down"`
	StartWeek int    `json:"start_week"`
}
