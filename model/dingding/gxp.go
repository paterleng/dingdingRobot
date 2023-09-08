package dingding

import (
	"crypto/tls"
	"ding/global"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TongXinUser struct {
	ID        string `gorm:"primarykey" json:"userid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string
	IsSchool  bool     `json:"is_school"` //是否留校考研
	Records   []Record //每个同学有多个记录
}
type Record struct {
	gorm.Model
	IsAtRobot     bool   `json:"is_at_robot"`      //是否@过机器人
	IsInRoom      bool   `json:"is_in_room"`       //是否回复了已到宿舍
	Content       string `json:"content"`          //回复的内容
	TongXinUserID string `json:"tong_xin_user_id"` //用户ID
}

// ImportUserToMysql 把钉钉用户信息导入到数据库中
func (d *TongXinUser) ImportUserToMysql() error {
	return global.GLOAB_DB1.Transaction(func(tx *gorm.DB) (err error) {
		token := "ae2372daaf6c3f4e829dff126d8770b0"
		Dept := &DingDept{DeptId: 1, DingToken: DingToken{Token: token}}
		DeptUserList, _, err := Dept.GetUserListByDepartmentID1(0, 100)
		fmt.Println(DeptUserList)
		if err != nil {
			zap.L().Error("获取部门成员失败", zap.Error(err))
		}
		tx.Create(&DeptUserList)
		if err != nil {
			zap.L().Error(fmt.Sprintf("存储部门:%s成员到数据库失败", Dept.Name), zap.Error(err))
		}
		return
	})
}
func (d *DingDept) GetUserListByDepartmentID1(cursor, size int) (userList []TongXinUser, hasMore bool, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/list?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
		Cursor int `json:"cursor"`
		Size   int `json:"size"`
	}{
		DeptID: d.DeptId,
		Cursor: cursor,
		Size:   size,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			HasMore bool          `json:"has_more"`
			List    []TongXinUser `json:"list"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, false, errors.New(r.Errmsg)
	}

	// 此处举行具体的逻辑判断，然后返回即可
	return r.Result.List, r.Result.HasMore, nil
}
