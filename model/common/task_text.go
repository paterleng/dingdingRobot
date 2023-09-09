package common

import "gorm.io/gorm"

type MsgText struct {
	gorm.Model
	At      At     `json:"at"`   //存储的是@的成员
	Text    Text   `json:"text"` //存储的是用户所发送的信息
	Msgtype string `json:"msgtype"`
	TaskID  uint   //MsgText属于task,我们次数使用的uint，而是没有使用foreignkey重写外键，所以此处指向的Task表中的Model中的ID
}
type At struct {
	gorm.Model
	//在数据库中遇到了数组类型的元素，其实就是一对多关系
	AtMobiles     []AtMobile `json:"atMobiles"` //At和Tele是一对多关系
	AtUserIds     []AtUserId `json:"atUserIds"`
	IsAtAll       bool       `json:"isAtAll"`
	MsgTextID     uint       //At属于MsgText,打上MsgText的标签
	MsgMarkDownID uint       //At也属于
}
type AtMobile struct {
	gorm.Model
	AtMobile string `json:"atMobile"`
	Name     string `json:"name"`
	AtID     uint   //AtMobile属于At，打上标签
}
type AtUserId struct {
	gorm.Model
	AtUserId string `json:"atUserId"`
	AtID     uint   //AtUserId属于At，打上标签
}
type Text struct {
	gorm.Model
	Content   string `json:"content"`
	MsgTextID uint   //Text属于At，打上标签
}
