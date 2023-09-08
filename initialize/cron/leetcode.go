package cron

import (
	"context"
	"ding/global"
	"ding/initialize/redis"
	"ding/model/common"
	"ding/model/dingding"
	"encoding/json"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SetWeek() (err error) {
	err = global.GLOBAL_REDIS.SetNX(context.Background(), "leetCode:week", 1, 0).Err()
	if err != nil {
		zap.L().Error("初始化week存入redis失败", zap.Error(err))
	}
	return
}
func SendLeetCode() (err error) {
	//每周一下午2点30运行
	spec := "0 30 14 ? * 1 "
	//spec := "10 33 14 ? * 3 "
	//spec:="0 0 0 ? * * "
	//开启定时器，定时周一晚上00：00(cron定时任务的创建)
	entryID, err := global.GLOAB_CORN.AddFunc(spec, func() {
		var week int
		//每次运行是redis中的week自增
		_, err = global.GLOBAL_REDIS.Incr(context.Background(), "leetCode:week").Result()
		if err != nil {
			zap.L().Error("redis中week自增失败", zap.Error(err))
			return
		}
		//从redis中读取week的值
		result, err := global.GLOBAL_REDIS.Get(context.Background(), "leetCode:week").Result()
		if err != nil {
			zap.L().Error("从redis中读取week失败", zap.Error(err))
			return
		}
		week, err = strconv.Atoi(result)
		if err != nil {
			zap.L().Error("week字符串转换失败", zap.Error(err))
		}

		//val, err := rdb.Incr(ctx, "key").Result()
		//if err != nil {
		//	panic(err)
		//}
		// 加载数据库部门表，找到需要查力扣的部门（gorm预加载https://gorm.io/zh_CN/docs/preload.html）
		var depts []dingding.DingDept
		err = global.GLOAB_DB.Where("is_leet_code=?", 1).Preload("UserList").Find(&depts).Error
		if err != nil || len(depts) == 0 {
			zap.L().Error("获取需要查询力扣的部门错误", zap.Error(err))
			return
		}

		//遍历部门
		for _, dept := range depts {
			zap.L().Info(fmt.Sprintf("%s开启了查询力扣题目,部门id是%d", dept.Name, dept.DeptId))
			//遍历某部门的同学，拿到力扣主页地址题目数据
			userList := dept.UserList
			t := time.Now().Format("2006-01-02")
			lastWeek := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
			weekDay := fmt.Sprintf("第%d周(%s)", week, t)
			oldDay := fmt.Sprintf("第%d周(%s)", week-1, lastWeek)
			oldNum := make(map[string]int)
			newNum := make(map[string]int)
			sub := make(map[string]int)
			for _, user := range userList {
				// 爬取本周数据，并存储
				leetCodeNum, err := getLeetCodeNum(user.LeetCodeAddr)
				if err != nil {
					zap.L().Error("获取用户力扣做题数错误", zap.Error(err))
				}
				key := fmt.Sprintf("%s%s:%s(%d):", redis.LeetCode, weekDay, dept.Name, dept.DeptId)
				err = global.GLOBAL_REDIS.ZAdd(context.Background(), key, &redis2.Z{Score: float64(leetCodeNum), Member: user.Name}).Err()
				if err != nil {
					zap.L().Error("获取做题数后存入redis失败", zap.Error(err))
				}
				// 对比两周数据
				username := user.Name
				newNum[username] = leetCodeNum
				key2 := fmt.Sprintf("%s%s:%s(%d):", redis.LeetCode, oldDay, dept.Name, dept.DeptId)
				num, err2 := global.GLOBAL_REDIS.ZScore(context.Background(), key2, user.Name).Result()
				if err2 != nil {
					zap.L().Error("读取该用户做题数失败", zap.Error(err2))
				}
				oldNum[username] = int(num)
				sub[username] = leetCodeNum - int(num)
			}
			message := weekDay + "力扣题目查询结果如下(该结果是通过查询总提交数来比较，如果您之前做过，那么不会增加总做题数)：\n" +
				"姓名-上周总题数-本周总提数-上周完成题目数\n"
			for name, _ := range oldNum {
				message += name + "-" + strconv.Itoa(oldNum[name]) + "-" + strconv.Itoa(newNum[name]) + "-" + strconv.Itoa(sub[name]) + "\n"
			}
			fmt.Println(message)

			zap.L().Info("message编辑完成，开始封装发送信息参数")
			p := &dingding.ParamCronTask{
				MsgText: &common.MsgText{
					At: common.At{
						IsAtAll: true,
					},
					Text: common.Text{
						Content: message,
					},
					Msgtype: "text",
				},
				RobotId: dept.RobotToken,
				//RobotId: "b14ef369d04a9bbfc10f3092d58f7214819b9daa93f3998121661ea0f9a80db3",
			}
			err := (&dingding.DingRobot{
				RobotId: dept.RobotToken,
				//RobotId: "b14ef369d04a9bbfc10f3092d58f7214819b9daa93f3998121661ea0f9a80db3",
			}).SendMessage(p)
			if err != nil {
				zap.L().Error("发送力扣题目消息失败", zap.Error(err))
				return
			}
		}

		return
	})
	if err != nil {
		zap.L().Error("力扣定时任务错误", zap.Error(err))
		return
	}
	fmt.Println("力扣定时任务", entryID)
	return
}

//func getUserByDeptId(deptId int) (users []dingding.DingUser, err error) {
//	//var userIDs map[string]userDept
//	var userIDs []string
//	err = global.GLOAB_DB.Table("user_dept").Where("ding_dept_dept_id", deptId).Select("ding_user_user_id").Find(&userIDs).Error
//	if err != nil {
//		zap.L().Error("获取部门用户失败", zap.Error(err))
//		return
//	}
//
//	//for _, v := range userIDs {
//	err = global.GLOAB_DB.Where("user_id in (?)", userIDs).Find(&users).Error
//	if err != nil {
//		zap.L().Error("通过userid查询user信息失败", zap.Error(err))
//	}
//	//}
//	return
//}

type Datas struct {
	Data struct {
		UserProfileUserQuestionProgress struct {
			NumAcceptedQuestions []struct {
				Difficulty string `json:"difficulty"`
				Count      int    `json:"count"`
			} `json:"numAcceptedQuestions"`
			NumFailedQuestions []struct {
				Difficulty string `json:"difficulty"`
				Count      int    `json:"count"`
			} `json:"numFailedQuestions"`
			NumUntouchedQuestions []struct {
				Difficulty string `json:"difficulty"`
				Count      int    `json:"count"`
			} `json:"numUntouchedQuestions"`
		} `json:"userProfileUserQuestionProgress"`
	} `json:"data"`
}

func getLeetCodeNum(leetCodeAddress string) (count int, err error) {
	url := "https://leetcode.cn/graphql/"
	method := "POST"
	payload := strings.NewReader(`{"query":"\n    query userQuestionProgress($userSlug: String!) {\n  userProfileUserQuestionProgress(userSlug: $userSlug) {\n    numAcceptedQuestions {\n      difficulty\n      count\n    }\n    numFailedQuestions {\n      difficulty\n      count\n    }\n    numUntouchedQuestions {\n      difficulty\n      count\n    }\n  }\n}\n    ","variables":{"userSlug":"` + leetCodeAddress + `"},"operationName":"userQuestionProgress"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Origin", "https://leetcode.cn")
	req.Header.Add("Referer", "https://leetcode.cn/u/mgy001/")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	req.Header.Add("authorization", "")
	req.Header.Add("baggage", "sentry-environment=production,sentry-release=PKr3xafTkhtDRjkc1yHvk,sentry-transaction=%2Fu%2F%5Busername%5D,sentry-public_key=7e9f5c528a9f4ee3b2bd215153cb69a7,sentry-trace_id=0b19d46a61864df7b52b9362a634f9b2,sentry-sample_rate=0.004")
	req.Header.Add("random-uuid", "7e09b8f9-f22e-a184-a6f2-73ab1023d23b")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sentry-trace", "0b19d46a61864df7b52b9362a634f9b2-a66907511bc1203d-0")
	req.Header.Add("x-csrftoken", "vXeFshgEi0fvllNmBSlFuUmK6g9wnayGieKJNMnavdcO9DQ4cniPP1u003AS5SG6")
	req.Header.Add("Cookie", "csrftoken=vXeFshgEi0fvllNmBSlFuUmK6g9wnayGieKJNMnavdcO9DQ4cniPP1u003AS5SG6; gr_user_id=54ba0057-ac8a-4e52-84ec-2ad5812054b0; a2873925c34ecbd2_gr_last_sent_cs1=mgy001; __atuvc=1%7C45%2C1%7C46%2C1%7C47; Hm_lvt_fa218a3ff7179639febdb15e372f411c=1677827295; _bl_uid=6mlqth5jbX5sRdfbn24nsyslkdze; gioCookie=yes; _gid=GA1.2.876186341.1683875161; NEW_QUESTION_DETAIL_PAGE_V2=1; Hm_lvt_f0faad39bcf8471e3ab3ef70125152c3=1683365606,1683640113,1683875161,1683895626; LEETCODE_SESSION=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJuZXh0X2FmdGVyX29hdXRoIjoiL3N1Ym1pc3Npb25zL2RldGFpbC80Mjg0MTY1MzQvIiwiX2F1dGhfdXNlcl9pZCI6IjQ3NDA1ODUiLCJfYXV0aF91c2VyX2JhY2tlbmQiOiJkamFuZ28uY29udHJpYi5hdXRoLmJhY2tlbmRzLk1vZGVsQmFja2VuZCIsIl9hdXRoX3VzZXJfaGFzaCI6ImJlZTlkMmExYmNmZjU1YjNlNTcwMjRjZjdkMDY5OWE2OWZjYzM2NmE1N2EwMTZiOWNkNWFmNjhmYWRkODA2N2QiLCJpZCI6NDc0MDU4NSwiZW1haWwiOiIiLCJ1c2VybmFtZSI6Im1neTAwMSIsInVzZXJfc2x1ZyI6Im1neTAwMSIsImF2YXRhciI6Imh0dHBzOi8vYXNzZXRzLmxlZXRjb2RlLmNuL2FsaXl1bi1sYy11cGxvYWQvdXNlcnMvb2pMampIN3F0ZS9hdmF0YXJfMTY0OTkyNDQ1OC5wbmciLCJwaG9uZV92ZXJpZmllZCI6dHJ1ZSwiX3RpbWVzdGFtcCI6MTY4MzY0MjY5Ny4yNjM5MjMsImV4cGlyZWRfdGltZV8iOjE2ODYxNjQ0MDAsInZlcnNpb25fa2V5XyI6MCwibGF0ZXN0X3RpbWVzdGFtcF8iOjE2ODM4OTU2MjV9.zpVkrvNJkDc86ppG7vA-xUgbblO1svM97HWO4v9u6CE; a2873925c34ecbd2_gr_session_id=fda95b3d-7872-4c9b-a48b-57bd23beda16; a2873925c34ecbd2_gr_last_sent_sid_with_cs1=fda95b3d-7872-4c9b-a48b-57bd23beda16; a2873925c34ecbd2_gr_session_id_sent_vst=fda95b3d-7872-4c9b-a48b-57bd23beda16; _ga_PDVPZYN3CW=GS1.1.1683895626.24.0.1683895626.0.0.0; _ga=GA1.2.300765497.1667698342; Hm_lpvt_f0faad39bcf8471e3ab3ef70125152c3=1683895634; a2873925c34ecbd2_gr_cs1=mgy001")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Host", "leetcode.cn")

	res, err := client.Do(req)
	if err != nil {
		zap.L().Error("获取力扣数据失败", zap.Error(err))
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		zap.L().Error("获取力扣数据失败", zap.Error(err))
		return
	}
	var userProfile Datas
	err = json.Unmarshal([]byte(string(body)), &userProfile)
	if err != nil {
		zap.L().Error("获取力扣数据失败", zap.Error(err))
		return
	}
	easy := userProfile.Data.UserProfileUserQuestionProgress.NumAcceptedQuestions[0].Count
	medium := userProfile.Data.UserProfileUserQuestionProgress.NumAcceptedQuestions[1].Count
	hard := userProfile.Data.UserProfileUserQuestionProgress.NumAcceptedQuestions[2].Count

	count = easy + medium + hard
	return
}
