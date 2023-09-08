package common

import "gorm.io/gorm"

type MsgLink struct {
	gorm.Model
	Msgtype string `json:"msgtype"`
	Link    Link   `json:"link"`
	TaskID  uint   //我们次数使用的uint，而是没有使用foreignkey重写外键，所以此处指向的Task表中的Model中的ID
}
type Link struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicUrl     string `json:"picUrl"`
	MessageUrl string `json:"messageUrl"`
	MsgLinkID  uint
}
