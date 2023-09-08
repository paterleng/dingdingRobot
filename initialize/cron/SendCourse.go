package cron

import (
	"crypto/tls"
	"ding/global"
	"ding/model/classCourse"
	"ding/model/dingding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func RegularlySendCourses() (err error) {
	var p dingding.ParamChat
	global.GLOAB_CORN.AddFunc("0 0 8 * * ?", func() {
		//查询所有订阅课表的结果
		sr, err := (&dingding.SubscriptionRelationship{}).QuerySubscribed()
		if err != nil {
			return
		}
		//遍历所有订阅课表的查询结果
		for _, value := range sr {
			var userids []string
			userids = append(userids, value.Subscriber)
			username, _ := (&dingding.DingUser{UserId: value.Subscribee}).GetUserByUserId()
			p.UserIds = userids
			p.MsgKey = "sampleText"
			p.MsgParam = fmt.Sprintf("姓名:%v\n", username.Name)

			//获取被订阅人的今日课程情况，并拼接到一起
			//获取当前是第几周
			week, err := (&classCourse.Calendar{}).GetWeek()
			if err != nil {
				return
			}

			//获取本周被订阅人的全部课程
			var client *http.Client
			client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}, Timeout: time.Duration(time.Second * 5)}
			URL := "http://localhost:20080/course/findCourseOfWeek?userid=" + value.Subscribee
			URL = URL + "&week=" + strconv.Itoa(week)
			request, err := http.NewRequest(http.MethodGet, URL, nil)
			if err != nil {
				return
			}
			resp, err := client.Do(request)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
			if err != nil {
				return
			}
			type course struct {
				Userid  string `json:"userid"`
				CName   string `json:"c_name"`
				CLesson int8   `json:"c_lesson"`
				CWeek   int8   `json:"c_week"`
			}
			r := struct {
				Code int      `json:"code"`
				Msg  string   `json:"msg"`
				Data []course `json:"data"`
			}{}
			//把请求到的结构反序列化到专门接受返回值的对象上面
			err = json.Unmarshal(body, &r)
			if err != nil {
				return
			}

			//获取今天是星期几
			now := time.Now()
			weekday := now.Weekday()
			if weekday == 0 {
				weekday = 7
			}
			//将今天被订阅人的课表拼接
			s1 := fmt.Sprintf("第一节：无课\n")
			s2 := fmt.Sprintf("第二节：无课\n")
			s3 := fmt.Sprintf("第三节：无课\n")
			s4 := fmt.Sprintf("第四节：无课\n")
			s5 := fmt.Sprintf("第五节：无课\n")
			for _, v := range r.Data {
				if int8(weekday) == v.CWeek {
					switch v.CLesson {
					case 1:
						s1 = fmt.Sprintf("第一节：%v\n", v.CName)
					case 2:
						s2 = fmt.Sprintf("第二节：%v\n", v.CName)
					case 3:
						s3 = fmt.Sprintf("第三节：%v\n", v.CName)
					case 4:
						s4 = fmt.Sprintf("第四节：%v\n", v.CName)
					case 5:
						s5 = fmt.Sprintf("第五节：%v\n", v.CName)
					default:
					}
				}
			}
			p.MsgParam = p.MsgParam + s1 + s2 + s3 + s4 + s5
			//发送被订阅人课程信息给订阅人
			token, _ := (&dingding.DingToken{}).GetAccessToken()
			_ = (&dingding.DingRobot{DingToken: dingding.DingToken{token}}).ChatSendMessage(&p)
		}
	})
	return
}
