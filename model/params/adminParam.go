package params

type GetUsers struct {
}
type SearchUser struct {
	UserName string `form:"username"`
}
type RemoveUser struct {
	UserId string `json:"user_id" binding:"required"`
}
