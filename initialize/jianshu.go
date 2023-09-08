package initialize

import (
	dingding2 "ding/model/dingding"
)

func JianBlogByRobot() (err error) {
	err = (&dingding2.DingUser{}).GoCrawlerDingUserJinAndBlog()
	return
}
