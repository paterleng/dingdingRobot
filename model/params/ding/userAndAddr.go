package ding

/**
*
* @author yth
* @language go
* @since 2023/2/16 20:10
 */

type UserAndAddrParam struct {
	UserId      string `json:"user_id" binding:"required"`
	JianShuAddr string `json:"jianshu_addr" binding:"required"`
	BlogAddr    string `json:"blog_addr" binding:"required"`
}
