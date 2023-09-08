package dingding

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type DingAttendance struct {
	UserCheckTime int64  `json:"userCheckTime"` //时间戳实际打卡时间。
	TimeResult    string `json:"timeResult"`    //打卡结果Normal：正常 Early：早退 Late：迟到 SeriousLate：严重迟到 Absenteeism：旷工迟到 NotSigned：未打卡
	CheckType     string `json:"checkType"`     //OnDuty 上班，OffDuty下班
	UserID        string `json:"userId"`
	UserName      string `json:"user_name"`
	DingToken
}

//获取考勤数据//获取考勤结果（可以根据userid批量查询） https://open.dingtalk.com/document/orgapp/attendance-clock-in-record-is-open
func (a *DingAttendance) GetAttendanceList(userIds []string, CheckDateFrom string, CheckDateTo string) (Recordresult []DingAttendance, err error) {
	zap.L().Info("进入到了获取数据的接口")
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/attendance/listRecord?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		CheckDateFrom string   `json:"checkDateFrom"`
		CheckDateTo   string   `json:"checkDateTo"`
		UserIds       []string `json:"userIds"`
	}{
		CheckDateFrom: CheckDateFrom,
		CheckDateTo:   CheckDateTo,
		UserIds:       userIds,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	zap.L().Info(fmt.Sprintf("把参数序列化到结构体对象上成功%v", bodymarshal))
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
	zap.L().Info(fmt.Sprintf("发送请求成功，原始resp为:%v", resp))
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Recordresult []DingAttendance `json:"recordresult"`
	}{}

	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	zap.L().Info(fmt.Sprintf("把请求结果序列化到结构体对象中成功%v", r))
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	return r.Recordresult, nil
}
