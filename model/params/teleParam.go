package params

type ParamAddTele struct {
	RobotId    string `json:"robot_id" binding:"required"`
	PersonName string `json:"person_name"`
	Number     string `json:"number"`
}
type ParamUpdateTele struct {
	Id            uint   `json:"id" binding:"required"`
	NewNumber     string `json:"new_number" binding:"required"`
	NewPersonName string `json:"new_person_name" binding:"required"`
}
type ParamRemoveTele struct {
	Id uint `json:"id" binding:"required"`
}

type ParamGetTeles struct {
	RobotId string `binding:"required" form:"robot_id"`
}
type ParamBatchInsertGroupMembers struct {
	//我们需要使用第三方接口来获取到chatId
	RobotID string   `json:"robot_id" binding:"required"`
	UserIds []string `json:"user_ids"` //群成员userid
}
